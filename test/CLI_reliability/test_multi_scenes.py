import os
import re
import sys
import logging
import time
import unittest

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import deleteDependentData
from common import checkServerStatus
from common import sysCommand
from common import getTuneTaskResult
from common import getTaskLogPath
from common import runSensitizeCollect

logger = logging.getLogger(__name__)


class TestMultiScenes(unittest.TestCase):
    @classmethod
    def setUpClass(self) -> None:
        self.brain = self.get_server_ip()
        if self.brain != "localhost":
            status = sysCommand("scp conf/restart_brain.sh {}:/opt".format(self.brain))[0]
            assert status == 0
        
    @classmethod
    def tearDownClass(self) -> None:
        if self.brain != "localhost":
            status = sysCommand("ssh {} 'rm -rf /opt/restart_brain.sh'".format(self.brain))[0]
            assert status == 0

    def setUp(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        logger.info('start to run test_multiple_scenes testcase')

    def tearDown(self) -> None:
        server_list = ["keentuned", "keentune-brain",
                       "keentune-target", "keentune-bench"]
        status = checkServerStatus(server_list)
        self.assertEqual(status, 0)
        deleteDependentData("param1")
        deleteDependentData("sensitize1")
        logger.info('the test_multiple_scenes testcase finished')

    @staticmethod
    def get_server_ip():
        with open("common.py", "r", encoding='UTF-8') as f:
            data = f.read()
        brain = re.search(r"brain_ip=\"(.*)\"", data).group(1)
        return brain

    def check_param_tune_job(self, name):
        cmd = 'keentune param jobs'
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__(name))

    def run_sensitize_train(self, name):
        cmd = "echo y | keentune sensitize train --data {} --output {}".format(name, name)
        self.status, self.out, _ = sysCommand(cmd)
        self.assertEqual(self.status, 0)

        path = re.search(r'\s+"(.*?)"', self.out).group(1)
        time.sleep(3)
        while True:
            with open(path, 'r') as f:
                res_data = f.read()
            if '"sensitize train" finish' in res_data:
                break
            time.sleep(8)

        word_list = ["Step1", "Step2", "Step3", "Step4", '"sensitize train" finish']
        result = all([word in res_data for word in word_list])
        self.assertTrue(result)

        self.path = "/var/keentune/sensitize/sensi-{}.json".format(name)
        res = os.path.exists(self.path)
        self.assertTrue(res)

    def restart_brain_server(self, algorithm):
        if self.brain == "localhost":
            cmd = "sh conf/restart_brain.sh {}".format(algorithm)
        else:
            cmd = "ssh {} 'sh /opt/restart_brain.sh {}'".format(self.brain, algorithm)
        
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertTrue(self.out.__contains__('restart brain server successfully!'))

    def test_param_tune_FUN_nginx(self):
        cmd = "sh conf/reset_keentuned.sh {} {}".format("param", "nginx.json")
        self.status, self.out, _  = sysCommand(cmd)
        self.assertEqual(self.status, 0)
        self.assertIn("restart keentuned server successfully!", self.out)
        cmd = 'keentune param tune -i 1 --job param1'
        path = getTaskLogPath(cmd)
        result = getTuneTaskResult(path)
        self.assertTrue(result)
        self.check_param_tune_job("param1")
    
    def test_sensitize_train_FUN_lasso(self):
        self.restart_brain_server("lasso")
        status = runSensitizeCollect("sensitize1")
        self.assertEqual(status, 0)
        self.run_sensitize_train("sensitize1")

    def test_sensitize_train_FUN_univariate(self):
        self.restart_brain_server("univariate")
        status = runSensitizeCollect("sensitize1")
        self.assertEqual(status, 0)
        self.run_sensitize_train("sensitize1")

    def test_sensitize_train_FUN_gp(self):
        self.restart_brain_server("gp")
        status = runSensitizeCollect("sensitize1")
        self.assertEqual(status, 0)
        self.run_sensitize_train("sensitize1")

    def test_sensitize_train_FUN_shap(self):
        self.restart_brain_server("shap")
        status = runSensitizeCollect("sensitize1")
        self.assertEqual(status, 0)
        self.run_sensitize_train("sensitize1")
