package media

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type boundedBuffer struct {
	mu    sync.Mutex
	data  bytes.Buffer
	limit int
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.data.Len()+len(p) > b.limit {
		return 0, errors.New("extractor output limit reached")
	}
	return b.data.Write(p)
}
func (b *boundedBuffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return bytes.Clone(b.data.Bytes())
}

func command(ctx context.Context, binary string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.WaitDelay = 3 * time.Second
	configureProcess(cmd)
	// No inherited browser/session credentials, proxy overrides or plugin paths.
	allowed := []string{"PATH=", "SYSTEMROOT=", "WINDIR=", "TEMP=", "TMP=", "LANG=", "HOME=", "DISPLAY="}
	for _, entry := range os.Environ() {
		for _, prefix := range allowed {
			if strings.HasPrefix(strings.ToUpper(entry), prefix) {
				cmd.Env = append(cmd.Env, entry)
				break
			}
		}
	}
	cmd.Env = append(cmd.Env, "PYTHONNOUSERSITE=1", "PYTHONUNBUFFERED=1")
	return cmd
}
