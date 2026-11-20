"""SimSight's Prism binding speaking through NVDA, under 64-bit Python 3.7 (as in the game)."""
import os, struct, sys, time
sys.path.insert(0, r'C:\SimSightTest\scripts')
out = open(r'C:\SimSightTest\prism-test.txt', 'w')
def log(m):
    out.write(m + '\n'); out.flush()
log('python %s, %d-bit' % (sys.version.split()[0], struct.calcsize('P') * 8))
try:
    from simsight_prism import Reader
    docs = os.path.join(os.environ['USERPROFILE'], 'Documents')
    dll = os.path.join(docs, 'Electronic Arts', 'The Sims 4', 'Mods', 'SimSight', 'SimSight_SpeechPack', 'prism.dll')
    r = Reader(dll, log)
    log('screen reader: %r' % (r.name,))
    log('output: %r' % (r.output('SimSight test. Hello from The Sims 4.', 'SimSight test', 'concise', True),))
    time.sleep(2)
    log('output 2: %r' % (r.output('Hunger low. 20 percent.', 'Hunger low', 'off', False),))
    time.sleep(3)
except Exception as ex:
    import traceback
    log('FAILED: ' + traceback.format_exc())
log('done')
