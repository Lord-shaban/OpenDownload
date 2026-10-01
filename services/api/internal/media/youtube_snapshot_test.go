package media

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestYouTubeSnapshotAvoidsRepeatedExtractionAndHidesCredentials(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix executable fixture; exercised by required Linux CI")
	}
	script := filepath.Join(t.TempDir(), "extractor")
	code := `#!/usr/bin/env python3
import json, pathlib, sys
args=sys.argv[1:]
calls=pathlib.Path(__file__+'.calls')
if '--dump-single-json' in args:
    with calls.open('a') as f: f.write('analyze\n')
    json.dump({'id':'test', 'title':'Public media', 'extractor_key':'Youtube',
        'webpage_url':'https://youtu.be/test', '__postprocessors':['unsafe'],
        'formats':[{'format_id':'18','ext':'mp4','height':360,'vcodec':'avc1','acodec':'mp4a','protocol':'https',
            'url':'https://cdn.example/media.mp4?signature=signed-test',
            'http_headers':{'User-Agent':'test','Cookie':'secret-cookie','Authorization':'secret-auth'}}]},sys.stdout)
else:
    assert '--load-info-json' in args and args[args.index('--load-info-json')+1]=='-'
    assert '--' not in args
    data=sys.stdin.read()
    assert 'webpage_url' not in data and '__postprocessors' not in data
    assert 'secret-cookie' not in data and 'secret-auth' not in data
    info=json.loads(data)
    assert info['formats'][0]['http_headers']['User-Agent']=='test'
    with calls.open('a') as f: f.write('download\n')
    (pathlib.Path(args[args.index('--paths')+1])/'media.mp4').write_bytes(b'test-output')
`
	if err := os.WriteFile(script, []byte(code), 0700); err != nil {
		t.Fatal(err)
	}
	e, err := NewYTDLP(script, "http://127.0.0.1:8090", 1024)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for i := 0; i < 2; i++ {
		a, err := e.Analyze(ctx, "https://youtu.be/test")
		if err != nil {
			t.Fatal(err)
		}
		public, _ := json.Marshal(a)
		if strings.Contains(string(public), "signed-test") || strings.Contains(string(public), "secret-") {
			t.Fatal("internal snapshot leaked into the API response")
		}
		if i == 1 {
			if _, err := e.Download(ctx, a, a.Options[0], t.TempDir(), func(Progress) {}); err != nil {
				t.Fatal(err)
			}
		}
	}
	calls, _ := os.ReadFile(script + ".calls")
	if string(calls) != "analyze\ndownload\n" {
		t.Fatalf("repeated extraction: %q", calls)
	}
}

func TestYouTubeSnapshotExpirySourceBindingAndCapacity(t *testing.T) {
	e := &YTDLP{}
	for i := 0; i < 40; i++ {
		source := strings.Repeat("x", i+1)
		e.rememberYouTubeSnapshot(source, Analysis{URL: source, snapshotSource: source, snapshotExpires: time.Now().Add(time.Minute), extractorSnapshot: []byte(`{"formats":[{"format_id":"18"}]}`)})
	}
	if _, ok := e.cachedYouTubeSnapshot("x"); ok || len(e.snapshots) != 32 {
		t.Fatal("snapshot admission is unbounded or retained an evicted source")
	}
	a, ok := e.cachedYouTubeSnapshot(strings.Repeat("x", 40))
	if !ok || !snapshotSupports(a, Option{Kind: "video", Selector: "18"}) {
		t.Fatal("valid issued format missing")
	}
	a.URL = "https://youtu.be/another"
	if snapshotSupports(a, Option{Kind: "video", Selector: "18"}) {
		t.Fatal("snapshot reused for another source")
	}
	for i := range e.snapshots {
		e.snapshots[i].analysis.snapshotExpires = time.Now().Add(-time.Second)
	}
	if _, ok := e.cachedYouTubeSnapshot(strings.Repeat("x", 40)); ok || len(e.snapshots) != 0 {
		t.Fatal("expired signed metadata was reused or retained")
	}
}
