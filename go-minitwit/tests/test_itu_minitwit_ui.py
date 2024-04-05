"""
To run this test with a visible browser, the following dependencies have to be setup:

  * `pip install selenium`
  * `pip install pymongo`
  * `pip install pytest`
  * `wget https://github.com/mozilla/geckodriver/releases/download/v0.32.0/geckodriver-v0.32.0-linux64.tar.gz`
  * `tar xzvf geckodriver-v0.32.0-linux64.tar.gz`
  * After extraction, the downloaded artifact can be removed: `rm geckodriver-v0.32.0-linux64.tar.gz`

The application that it tests is the version of _ITU-MiniTwit_ that you got to know during the exercises on Docker:
https://github.com/itu-devops/flask-minitwit-mongodb/tree/Containerize (*OBS*: branch Containerize)

```bash
$ git clone https://github.com/HelgeCPH/flask-minitwit-mongodb.git
$ cd flask-minitwit-mongodb
$ git switch Containerize
```

After editing the `docker-compose.yml` file file where you replace `youruser` with your respective username, the
application can be started with `docker-compose up`.

Now, the test itself can be executed via: `pytest test_itu_minitwit_ui.py`.
"""

import pymongo
import sqlite3
import pytest
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.firefox.service import Service
from selenium.webdriver.firefox.options import Options


GUI_URL = "http://localhost:8081/register"
#DB_URL = "mongodb://localhost:27017/test"
DB_URL = "../tmp/minitwit_empty.db"

def _register_user_via_gui(driver, data):
    import time    
    driver.get(GUI_URL)

    wait = WebDriverWait(driver, 5)
    buttons = wait.until(EC.presence_of_all_elements_located((By.CLASS_NAME, "actions")))
    input_fields = driver.find_elements(By.TAG_NAME, "input")

    for idx, str_content in enumerate(data):
        input_fields[idx].send_keys(str_content)
    input_fields[4].send_keys(Keys.RETURN)
    wait = WebDriverWait(driver, 5)

    def get_text_from_first_li(driver):
        try:
            flashes_ul = driver.find_element(By.CLASS_NAME, "flashes")
            li_elements = flashes_ul.find_elements(By.TAG_NAME, "li")
            if li_elements and li_elements[0].text.strip():
                return li_elements[0].text.strip()
        except:
            return None

    wait = WebDriverWait(driver, 5)
    li_text = wait.until(get_text_from_first_li)
    return li_text


def _get_user_by_name(db_client, name):
    return db_client.execute(f"SELECT username FROM user WHERE username='{name}'").fetchone()


def test_register_user_via_gui():
    """
    This is a UI test. It only interacts with the UI that is rendered in the browser and checks that visual
    responses that users observe are displayed.
    """
    firefox_options = Options()
    firefox_options.add_argument("--headless") # for visibility
    with webdriver.Firefox(options=firefox_options) as driver:
        generated_msg = _register_user_via_gui(driver, ["Me", "me@some.where", "secure123", "secure123"])
        expected_msg = "You were successfully registered and can login now"
        assert generated_msg == expected_msg

def test_register_user_via_gui_and_check_db_entry():
    """
    This is an end-to-end test. Before registering a user via the UI, it checks that no such user exists in the
    database yet. After registering a user, it checks that the respective user appears in the database.
    """
    firefox_options = Options()
    firefox_options.add_argument("--headless") # for visibility
    with webdriver.Firefox(options=firefox_options) as driver:
        con = sqlite3.connect(DB_URL)


        assert _get_user_by_name(con.cursor(), "Me") == None

        generated_msg = _register_user_via_gui(driver, ["Me", "me@some.where", "secure123", "secure123"])
        expected_msg = "You were successfully registered and can login now"
        assert generated_msg == expected_msg

        assert _get_user_by_name(con.cursor(), "Me")[0] == "Me"


@pytest.fixture(autouse=True)
def cleanupfix():
    con = sqlite3.connect(DB_URL)
    cur = con.cursor()
    cur.execute(f"DELETE FROM USER")
    con.commit()
