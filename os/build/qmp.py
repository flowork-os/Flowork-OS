#!/usr/bin/env python3
"""Tiny QMP client: negotiate then run one command. Used to screendump the VM
framebuffer to PNG for headless kiosk verification (qemu >= 7.1 supports PNG)."""
import json
import socket
import sys
import time


def qmp(sock_path, command, arguments=None, retries=30):
    last = None
    for _ in range(retries):
        try:
            s = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
            s.connect(sock_path)
            f = s.makefile("rw")
            f.readline()                       # greeting
            f.write(json.dumps({"execute": "qmp_capabilities"}) + "\n"); f.flush()
            f.readline()
            req = {"execute": command}
            if arguments:
                req["arguments"] = arguments
            f.write(json.dumps(req) + "\n"); f.flush()
            print(f.readline().strip())
            s.close()
            return 0
        except (FileNotFoundError, ConnectionRefusedError) as e:
            last = e
            time.sleep(1)
    print(f"qmp: could not reach {sock_path}: {last}", file=sys.stderr)
    return 1


if __name__ == "__main__":
    sock = sys.argv[1]
    cmd = sys.argv[2]
    args = json.loads(sys.argv[3]) if len(sys.argv) > 3 else None
    sys.exit(qmp(sock, cmd, args))
