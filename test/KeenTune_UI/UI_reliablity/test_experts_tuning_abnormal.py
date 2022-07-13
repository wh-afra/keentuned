import sys
import unittest
from time import sleep
from selenium import webdriver
from selenium.common.exceptions import NoSuchElementException
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.common.by import By


class TestKeenTune_UI_abnormal(unittest.TestCase):
    @classmethod
    def setUpClass(self,no_ui=False) -> None:
        if 'linux' in sys.platform:
            option = webdriver.ChromeOptions()
            option.add_argument('headless')
            option.add_argument('no-sandbox')
            option.add_argument('--start-maximized')
            option.add_argument('--disable-gpu')
            option.add_argument('--window-size=1920,1080')
            self.driver = webdriver.Chrome(options=option)
            self.driver.implicitly_wait(3)

        else:
            if no_ui:
                option = webdriver.ChromeOptions()
                option.add_argument('headless')
                option.add_argument('--start-maximized')
                self.driver = webdriver.Chrome(chrome_options=option)
            else:
                self.driver = webdriver.Chrome()
                self.driver.maximize_window()

        self.driver.get("http://39.102.53.144:8082/list/static-page")
        self.driver.find_element(By.XPATH,
                                 '//button[@class="ant-btn ant-btn-primary ant-btn-two-chinese-chars"]').click()
        self.driver.find_element(By.ID, "name").send_keys("1")
        self.driver.find_element(By.ID, "info").send_keys("[my.con]")
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]').click()

        self.driver.find_element(By.XPATH,
                                 '//button[@class="ant-btn ant-btn-primary ant-btn-two-chinese-chars"]').click()
        self.driver.find_element(By.ID, "name").send_keys("11")
        self.driver.find_element(By.ID, "info").send_keys("[my.con]")
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]').click()
        return self.driver

    @classmethod
    def tearDownClass(self) -> None:
        self.driver.find_element(By.XPATH, '//tr[@data-row-key="1"]//td[4]//div[1]//div[1]').click()
        sleep(1)
        self.driver.find_element(By.XPATH, '//div[@class="ant-popover-buttons"]/button[2]').click()

        self.driver.find_element(By.XPATH, '//tr[@data-row-key="1"]//td[4]//div[1]//div[1]').click()
        sleep(1)
        self.driver.find_element(By.XPATH, '//div[@class="ant-popover-buttons"]/button[2]').click()
        self.driver.quit()

    def setUp(self) -> None:
        try:
            self.driver.find_element(By.XPATH,
                                     '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]')
        except NoSuchElementException:
            pass

        else:
            self.driver.find_element(By.XPATH,
                                     '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]').click()


    def test_group_empty(self):
        self.driver.find_element(By.XPATH, '//tr[@data-row-key="1"]//td[4]//div[5]').click()
        sleep(1)
        self.driver.find_element(By.XPATH, '//div[3]/div/div[2]/button/span').click()
        ele_group_error = self.driver.find_element(By.XPATH,'//div[@class="ant-message-notice-content"]//span[2]')
        sleep(1)
        assert "请选择一个配置，再提交" in ele_group_error.text
        self.driver.find_element(By.XPATH,'//div[3]/div/div[1]/button/span').click()

    def test_copyfile_name_empty(self):
        self.driver.find_element(By.XPATH, '//tr[@data-row-key="1"]/td[4]//div[2]').click()
        self.driver.find_element(By.ID, "name").send_keys(Keys.CONTROL, "a")
        self.driver.find_element(By.ID, "name").send_keys(Keys.BACKSPACE)
        sleep(1)
        ele_emptyname = self.driver.find_element(By.XPATH,
                                                 '//div[@class="ant-form-item-explain ant-form-item-explain-connected"]')
        assert "请输入" in ele_emptyname.text
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]').click()

    def test_copyfile_context_empty(self):
        self.driver.find_element(By.XPATH, '//tr[@data-row-key="1"]/td[4]//div[2]').click()
        self.driver.find_element(By.ID, "info").send_keys(Keys.CONTROL, "a")
        self.driver.find_element(By.ID, "info").send_keys(Keys.BACKSPACE)
        sleep(1)
        ele_copy_context_empty = self.driver.find_element(By.XPATH,
                                                 '//div[@class="ant-form-item-explain ant-form-item-explain-connected"]')

        assert "请输入" in ele_copy_context_empty.text
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]').click()

    def test_copyfile_context_error(self):
        self.driver.find_element(By.XPATH, '//tr[@data-row-key="1"]/td[4]//div[2]').click()
        self.driver.find_element(By.ID, "info").send_keys(Keys.CONTROL, "a")
        self.driver.find_element(By.ID, "info").send_keys(Keys.BACKSPACE)
        self.driver.find_element(By.ID, "info").send_keys("error_file_context")
        sleep(1)
        ele_copy_context_error = self.driver.find_element(By.XPATH,
                                                          '//div[@class="ant-form-item-explain ant-form-item-explain-connected"]')
        assert "第 1 行数据格式不对!" in ele_copy_context_error.text
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]').click()

    def test_copyfile_name_exit(self):
        self.driver.find_element(By.XPATH, '//tr[@data-row-key="1"]/td[4]//div[2]').click()
        self.driver.find_element(By.ID, "name").send_keys(Keys.CONTROL, "a")
        self.driver.find_element(By.ID, "name").send_keys(Keys.BACKSPACE)
        self.driver.find_element(By.ID, "name").send_keys("1")
        sleep(1)
        ele_nameexit = self.driver.find_element(By.XPATH, '//div[@class="ant-form-item-explain-error"]')
        assert "Profile Name名字重复!" in ele_nameexit.text
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]').click()

    def test_creatfile_name_empty(self):
        self.driver.find_element(By.XPATH,
                                 '//button[@class="ant-btn ant-btn-primary ant-btn-two-chinese-chars"]').click()
        self.driver.find_element(By.ID, "info").send_keys("[my.con]")
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]').click()
        sleep(1)
        ele_nameempty = self.driver.find_element(By.XPATH, '//div[@class="ant-form-item-explain-error"]')
        assert "请输入" in ele_nameempty.text
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]').click()

    def test_creatfile_content_empty(self):
        self.driver.find_element(By.XPATH,
                                 '//button[@class="ant-btn ant-btn-primary ant-btn-two-chinese-chars"]').click()
        self.driver.find_element(By.ID, "name").send_keys("content_empty")
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]').click()
        sleep(1)
        ele_contentempty = self.driver.find_element(By.XPATH,
                                                    '//div[@class="ant-form-item-explain ant-form-item-explain-connected"]')
        assert "请输入" in ele_contentempty.text
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]').click()

    def test_creatfile_name_exit(self):
        self.driver.find_element(By.XPATH,
                                 '//button[@class="ant-btn ant-btn-primary ant-btn-two-chinese-chars"]').click()
        self.driver.find_element(By.ID, "name").send_keys("1")
        sleep(1)
        ele_nameexit = self.driver.find_element(By.XPATH,
                                                    '//div[@class="ant-form-item-explain-error"]')
        assert "Profile Name名字重复!" in ele_nameexit.text
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]').click()

    def test_creatfile_content_error(self):
        self.driver.find_element(By.XPATH,
                                 '//button[@class="ant-btn ant-btn-primary ant-btn-two-chinese-chars"]').click()
        self.driver.find_element(By.ID, "name").send_keys("content_error")
        self.driver.find_element(By.ID, "info").send_keys("content_error")
        sleep(1)
        ele_contenterror = self.driver.find_element(By.XPATH,
                                                    '//div[@class="ant-form-item-explain-error"]')
        assert "第 1 行数据格式不对!" in ele_contenterror.text
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]').click()

    def test_editorfile_delete_name(self):
        self.driver.find_element(By.XPATH, '//tr[@data-row-key="1"]/td[4]//div[3]').click()
        self.driver.find_element(By.ID, "name").send_keys(Keys.CONTROL, "a")
        self.driver.find_element(By.ID, "name").send_keys(Keys.BACKSPACE)
        ele_deletename = self.driver.find_element(By.XPATH,
                                                    '//div[@class="ant-form-item-explain-error"]')
        sleep(1)
        assert "请输入" in ele_deletename.text
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]').click()

    def test_editorfile_delete_content(self):
        self.driver.find_element(By.XPATH, '//tr[@data-row-key="1"]/td[4]//div[3]').click()
        self.driver.find_element(By.ID, "info").send_keys(Keys.CONTROL, "a")
        self.driver.find_element(By.ID, "info").send_keys(Keys.BACKSPACE)
        ele_deletecontent = self.driver.find_element(By.XPATH,
                                                    '//div[@class="ant-form-item-explain-error"]')
        sleep(1)
        assert "请输入" in ele_deletecontent.text
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]').click()

    def test_editorfile_exit_name(self):
        self.driver.find_element(By.XPATH, '//tr[@data-row-key="1"]/td[4]//div[3]').click()
        self.driver.find_element(By.ID, "name").send_keys(Keys.CONTROL, "a")
        self.driver.find_element(By.ID, "name").send_keys(Keys.BACKSPACE)
        self.driver.find_element(By.ID, "name").send_keys("11")
        ele_deletecontent = self.driver.find_element(By.XPATH,
                                                     '//div[@class="ant-form-item-explain-error"]')
        sleep(1)
        assert "Profile Name名字重复!" in ele_deletecontent.text
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]').click()

    def test_editorfile_error_content(self):
        self.driver.find_element(By.XPATH, '//tr[@data-row-key="1"]/td[4]//div[3]').click()
        self.driver.find_element(By.ID, "info").send_keys(Keys.CONTROL, "a")
        self.driver.find_element(By.ID, "info").send_keys(Keys.BACKSPACE)
        self.driver.find_element(By.ID, "info").send_keys("error_content")
        ele_errorcontent = self.driver.find_element(By.XPATH,'//div[@class="ant-form-item-explain-error"]')
        sleep(1)
        assert "第 1 行数据格式不对!" in ele_errorcontent.text
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]').click()