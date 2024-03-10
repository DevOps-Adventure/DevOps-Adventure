import os
import json
import base64
import sqlite3
import requests
import pytest
from pathlib import Path
from contextlib import closing


BASE_URL = "http://localhost:8081/api"
if os.getenv('EXECUTION_ENVIRONMENT') == 'CI':
    DATABASE = "../tmp/minitwit_empty.db"
    PATH_SCHEMA = "../schema.sql"
else:
    PATH_SCHEMA = "./schema.sql"
    DATABASE = "./tmp/minitwit_empty.db"

USERNAME = 'simulator'
PWD = 'super_safe!'
CREDENTIALS = ':'.join([USERNAME, PWD]).encode('ascii')
ENCODED_CREDENTIALS = base64.b64encode(CREDENTIALS).decode()
HEADERS = {'Connection': 'close',
           'Content-Type': 'application/json',
           f'Authorization': f'Basic {ENCODED_CREDENTIALS}'}



def init_db():
    """Creates the database tables."""
    with closing(sqlite3.connect(DATABASE)) as db:
        with open(PATH_SCHEMA) as fp:
            db.cursor().executescript(fp.read())
        db.commit()

# Empty the database and initialize the schema again
Path(DATABASE).unlink()
init_db()

def delete_user(username):
    # Assuming DATABASE points to your test database
    conn = sqlite3.connect(DATABASE)
    cur = conn.cursor()
    cur.execute("DELETE FROM user WHERE username = ?", (username,))
    # print results
    print(f"Deleted {cur.rowcount} user(s) with username {username}")
    conn.commit()
    conn.close()


@pytest.fixture(scope="session", autouse=True)
def clean_up_user(request):
    # Define cleanup function to run after all tests
    def remove_test_users():
        delete_user('a')
        delete_user('b')
        delete_user('c')
        print("Cleanup: Deleted test users")

    # finalizer to be run after all tests are done
    request.addfinalizer(remove_test_users)

def create_new_session():
    session = requests.Session()
    session.headers.update({
        'Connection': 'close',
        'Content-Type': 'application/json',
        'Authorization': f'Basic {ENCODED_CREDENTIALS}'
    })
    return session


def test_latest():
    session = create_new_session()
    # post something to update LATEST
    url = f"{BASE_URL}/register"
    data = {'username': 'test', 'email': 'test@test', 'pwd': 'foo'}
    params = {'latest': 1337}
    response = session.post(url, data=json.dumps(data),
                             params=params)
    assert response.ok

    # verify that latest was updated
    url = f'{BASE_URL}/latest'
    response = session.get(url)
    assert response.ok
    assert response.json()['latest'] == 1337


def test_register():
    session = create_new_session()
    username = 'a'
    email = 'a@a.a'
    pwd = 'a'
    data = {'username': username, 'email': email, 'pwd': pwd}
    params = {'latest': 1}
    response = session.post(f'{BASE_URL}/register',
                             data=json.dumps(data), params=params)
    assert response.ok
    # TODO: add another assertion that it is really there

    # verify that latest was updated
    response = session.get(f'{BASE_URL}/latest')
    assert response.json()['latest'] == 1


def test_create_msg():
    session = create_new_session()
    username = 'a'
    data = {'content': 'Blub!'}
    url = f'{BASE_URL}/msgs/{username}'
    params = {'latest': 2}
    response = session.post(url, data=json.dumps(data),
                             params=params)
    assert response.ok

    # verify that latest was updated
    response = session.get(f'{BASE_URL}/latest')
    assert response.json()['latest'] == 2


def test_get_latest_user_msgs():
    session = create_new_session()
    username = 'a'
    query = {'no': 20, 'latest': 3}
    url = f'{BASE_URL}/msgs/{username}'
    response = session.get(url, params=query)
    assert response.status_code == 200

    got_it_earlier = False
    for msg in response.json():
        if msg['content'] == 'Blub!' and msg['user'] == username:
            got_it_earlier = True

    assert got_it_earlier

    # verify that latest was updated
    response = session.get(f'{BASE_URL}/latest')
    assert response.json()['latest'] == 3


def test_get_latest_msgs():
    session =  create_new_session()
    username = 'a'
    query = {'no': 20, 'latest': 4}
    url = f'{BASE_URL}/msgs'
    response = session.get(url, params=query)
    assert response.status_code == 200

    got_it_earlier = False
    for msg in response.json():
        if msg['content'] == 'Blub!' and msg['user'] == username:
            got_it_earlier = True

    assert got_it_earlier

    # verify that latest was updated
    response = session.get(f'{BASE_URL}/latest')
    assert response.json()['latest'] == 4


def test_register_b():
    session =  create_new_session()
    username = 'b'
    email = 'b@b.b'
    pwd = 'b'
    data = {'username': username, 'email': email, 'pwd': pwd}
    params = {'latest': 5}
    response = session.post(f'{BASE_URL}/register', data=json.dumps(data),
                            params=params)
    assert response.ok
    # TODO: add another assertion that it is really there

    # verify that latest was updated
    response = session.get(f'{BASE_URL}/latest')
    assert response.json()['latest'] == 5


def test_register_c():
    session = create_new_session()
    username = 'c'
    email = 'c@c.c'
    pwd = 'c'
    data = {'username': username, 'email': email, 'pwd': pwd}
    params = {'latest': 6}
    response = session.post(f'{BASE_URL}/register', data=json.dumps(data),
                             params=params)
    assert response.ok

    # verify that latest was updated
    response = session.get(f'{BASE_URL}/latest')
    assert response.json()['latest'] == 6


def test_follow_user():
    session = create_new_session()
    username = 'a'
    url = f'{BASE_URL}/fllws/{username}'
    data = {'follow': 'b'}
    params = {'latest': 7}
    response = session.post(url, data=json.dumps(data),
                            params=params)
    assert response.ok

    data = {'follow': 'c'}
    params = {'latest': 8}
    response = session.post(url, data=json.dumps(data),
                            params=params)
    assert response.ok

    query = {'no': 20, 'latest': 9}
    response = session.get(url, params=query)
    assert response.ok

    json_data = response.json()
    assert "b" in json_data["follows"]
    assert "c" in json_data["follows"]

    # verify that latest was updated
    response = session.get(f'{BASE_URL}/latest')
    assert response.json()['latest'] == 9


def test_a_unfollows_b():
    session = create_new_session()
    username = 'a'
    url = f'{BASE_URL}/fllws/{username}'

    #  first send unfollow command
    data = {'unfollow': 'b'}
    params = {'latest': 10}
    response = session.post(url, data=json.dumps(data),
                            params=params)
    assert response.ok

    # then verify that b is no longer in follows list
    query = {'no': 20, 'latest': 11}
    response = session.get(url, params=query)
    assert response.ok
    assert 'b' not in response.json()['follows']

    # verify that latest was updated
    response = session.get(f'{BASE_URL}/latest')
    assert response.json()['latest'] == 11