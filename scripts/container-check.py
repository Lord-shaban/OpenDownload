"""Linux Compose smoke checks. Fails if the API has direct outbound internet."""
import json
import subprocess
import urllib.request

with urllib.request.urlopen("http://localhost:3000/api/v1/status", timeout=10) as response:
    status = json.load(response)
assert status["ready"] and status["fixtureMode"], status
assert all(status["dependencies"].values()), status

direct = """
import socket
try:
    connection = socket.create_connection(('1.1.1.1',443),timeout=3)
except OSError:
    print('Direct extractor egress is blocked.')
else:
    connection.close()
    raise SystemExit('FAIL: API has direct internet egress')
"""
blocked = """
import urllib.request, urllib.error
opener=urllib.request.build_opener(urllib.request.ProxyHandler({'http':'http://egress:8090','https':'http://egress:8090'}))
for target in ('http://169.254.169.254/latest/meta-data/','https://127.0.0.1/'):
    try:
        opener.open(target,timeout=5)
    except (urllib.error.URLError, OSError):
        print('Private destination refused.')
    else:
        raise SystemExit('FAIL: private destination was reachable')
"""
for code in (direct, blocked):
    subprocess.run(["docker", "compose", "exec", "-T", "api", "python3", "-c", code], check=True)
web_direct = """
const socket = require('node:net').connect({host:'1.1.1.1',port:443});
socket.setTimeout(3000);
socket.on('connect',()=>{console.error('FAIL: web has direct internet egress');socket.destroy();process.exit(1)});
socket.on('error',()=>{console.log('Web egress blocked.');process.exit(0)});
socket.on('timeout',()=>{socket.destroy();console.log('Web egress blocked.');process.exit(0)});
"""
subprocess.run(["docker", "compose", "exec", "-T", "web", "node", "-e", web_direct], check=True)
print("Compose smoke checks passed.")
