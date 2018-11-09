#!/usr/bin/env python2

import lbc
import re
import collections
import tempfile
import shutil
import unittest

# Exception that gets thrown when unexpected command is
# executed by lbc script.
class UnexpectedCmdException(Exception):
    def __init__(self, *args, **kwargs):
        Exception.__init__(self, *args, **kwargs)

# Exception that will get thrown in place of sys.exit()
class TestFailException(Exception):
    def __init__(self, *args, **kwargs):
        Exception.__init__(self, *args, **kwargs)

def test_fail(msg):
    raise TestFailException(msg)

# A FIFO queue of expected commands will be kept here
expected_cmds = collections.deque()

# Adds a command to the list
def expect_cmd(re_pattern, returncode=0, stdout=''):
    expected_cmds.append((re_pattern, returncode, stdout))

def test_run(cmd, timeout=None, stdin=None, show_stderr=True):
    cmd_re, returncode, stdout = expected_cmds.popleft()
    if re.match(cmd_re, cmd) != None:
        return stdout, returncode
    raise UnexpectedCmdException("Expected a command that matches '{}', instead got '{}'"
                                 .format(cmd_re, cmd))

def test_print(*args, **kwargs):
    pass

tempdir = ''
def test_make_tempdir():
    return tempdir

class LbcTest(unittest.TestCase):
    def setUp(self):
        # Create tempdir & fake chartfile there for export yaml tests
        global tempdir
        tempdir = tempfile.mkdtemp()
        open(tempdir + '/enterprise-suite-1.0.0-rc.4.tgz', 'w').close()

        # Create another tempdir & credentials file
        self.creds_dir = tempfile.mkdtemp()
        self.creds_file = self.creds_dir + '/commercial.credentials'
        with open(self.creds_file, 'w') as creds:
            creds.write('realm = Bintray\nhost = dl.bintray.com\nuser = aaa\npassword = bbb\n')

    def tearDown(self):
        # Temp dir for test_make_tempdir will be removed by lbc itself

        # Remove credentials tempdir
        shutil.rmtree(self.creds_dir)

    def test_export_yaml_console(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch -d \S* --devel es-repo/enterprise-suite')
        expect_cmd(r'helm template --name enterprise-suite --namespace lightbend   \S*\.tgz')
        lbc.main(['install', '--skip-checks', '--export-yaml=console'])
        self.assertEqual(len(expected_cmds), 0)
    
    def test_export_yaml_creds(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch -d \S* --devel es-repo/enterprise-suite')
        expect_cmd(r'helm template --name enterprise-suite --namespace lightbend  --execute templates/commercial-credentials\.yaml --values \S+ \S+\.tgz')
        lbc.main(['install', '--skip-checks', '--creds='+self.creds_file, '--export-yaml=creds'])
        self.assertEqual(len(expected_cmds), 0)

    def test_install(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend --devel --values \S+')
        lbc.main(['install', '--skip-checks', '--creds='+self.creds_file])
        self.assertEqual(len(expected_cmds), 0)

    def test_upgrade(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm status enterprise-suite', returncode=0)
        expect_cmd(r'helm upgrade enterprise-suite es-repo/enterprise-suite --devel --values \S+ ')
        lbc.main(['install', '--skip-checks', '--creds='+self.creds_file])
        self.assertEqual(len(expected_cmds), 0)

    def test_force_install(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm status enterprise-suite', returncode=0)
        expect_cmd(r'helm delete --purge enterprise-suite')
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend --devel --values \S+')
        lbc.main(['install', '--skip-checks', '--creds='+self.creds_file, '--force-install'])
        self.assertEqual(len(expected_cmds), 0)

    def test_helm_args(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend --devel --values \S+ --set minikube=true --fakearg')
        lbc.main(['install', '--skip-checks', '--creds='+self.creds_file, '--', '--set', 'minikube=true', '--fakearg'])
        self.assertEqual(len(expected_cmds), 0)

    def test_specify_version(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend --version 1\.0\.0-rc\.9 --values \S+')
        lbc.main(['install', '--skip-checks', '--creds='+self.creds_file, '--version=1.0.0-rc.9'])
        self.assertEqual(len(expected_cmds), 0)

if __name__ == '__main__':
    # Override internal functions so we can test them
    lbc.run = test_run
    lbc.fail = test_fail
    lbc.printerr = test_print
    lbc.printinfo = test_print
    lbc.make_tempdir = test_make_tempdir

    # Run tests
    unittest.main()