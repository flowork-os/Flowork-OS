"""py-hello — a sandboxed Python sample app for Flowork OS (P3a).
Proves: a Python app runs under bwrap AND cannot reach the owner's state."""
import os

print("py-hello: ran ok (cwd=%s)" % os.getcwd())

# Isolation probe: the owner's state must be invisible inside the sandbox.
target = "/root/.flowork/flowork.db"
try:
    with open(target, "rb") as f:
        f.read(1)
    print("ISOLATION_BREACH: read", target)
except Exception as e:
    print("ISOLATION_OK: cannot reach %s (%s)" % (target, type(e).__name__))
