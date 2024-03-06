# -*- coding: utf-8 -*-
"""
    MiniTwit Tests
    ~~~~~~~~~~~~~~

    Tests a MiniTwit application.

    :refactored: (c) 2024 by HelgeCPH from Armin Ronacher's original unittest version
    :copyright: (c) 2010 by Armin Ronacher.
    :license: BSD, see LICENSE for more details.
"""
import requests
import pytest
import sqlite3
import os

# import schema
# import data
# otherwise use the database that you got previously
BASE_URL = "http://localhost:8081"
if os.getenv('EXECUTION_ENVIRONMENT') == 'CI':
    DATABASE = "../tmp/minitwit_empty.db"
else:
    DATABASE = "./tmp/minitwit_empty.db"

print("Running tests...")

def delete_user(username):
    # Assuming DATABASE points to your test database
    conn = sqlite3.connect(DATABASE)
    cur = conn.cursor()
    cur.execute("DELETE FROM user WHERE username = ?", (username,))
    # print results
    print(f"Deleted {cur.rowcount} user(s) with username {username}")
    conn.commit()
    conn.close()

@pytest.fixture(autouse=True)
def clean_up_user():
    # Setup code: Delete the test user if exists
    # This requires implementing a function that deletes a user by username from your database
    delete_user('user1')
    delete_user('meh')
    # No teardown code needed if you're cleaning up before each test
    yield

@pytest.fixture(scope="session", autouse=True)
def clean_up_user(request):
    # Define cleanup function to run after all tests
    def remove_test_users():
        delete_user('user1')
        delete_user('meh')
        delete_user('foo')
        delete_user('bar')
        print("Cleanup: Deleted test users")

    # finalizer to be run after all tests are done
    request.addfinalizer(remove_test_users)

def register(username, password, passwordConfirm=None, email=None, session=None, follow_redirects=True):
    """Helper function to register a user"""
    if session is None:
        # raise ValueError("Session object must be provided")
        session = requests.Session()
    if passwordConfirm is None:
        passwordConfirm = password  # Use the same password if passwordConfirm isn't provided
    if email is None:
        email = f'{username}@example.com'  # Construct a default email if one isn't provided
    
    response = session.post(f'{BASE_URL}/register', data={
        'username': username,
        'password': password,
        'passwordConfirm': passwordConfirm,
        'email': email,
    }, allow_redirects=follow_redirects)
    
    # Check for HTTP errors (4xx, 5xx) after the request is made
    try:
        response.raise_for_status()
    except requests.exceptions.HTTPError as e:
        # Optionally, log the error or handle it based on your test requirements
        print(f"HTTP Error occurred: {e.response.status_code} - {e.response.reason}")
    
    return response

def login(username, password):
    """Helper function to login"""
    http_session = requests.Session()
    r = http_session.post(f'{BASE_URL}/login', data={
        'username': username,
        'password': password
    }, allow_redirects=True)
    return r, http_session

def register_and_login(username, password):
    """Registers and logs in in one go"""
    register(username, password)
    return login(username, password)

def logout(http_session):
    """Helper function to logout"""
    return http_session.get(f'{BASE_URL}/logout', allow_redirects=True)

def add_message(http_session, text):
    """Records a message"""
    r = http_session.post(f'{BASE_URL}/add_message', data={'text': text},
                                allow_redirects=True)
    if text:
        assert 'Your message was recorded' in r.text
    return r

# testing functions

def test_register():
    """Make sure registering works"""
    with requests.Session() as session:
        # Assuming delete_user() correctly resets the state before this test
        r = register('user1', 'default', session=session)
        assert 'You were successfully registered and can login now' in r.text
        r = register('user1', 'default')
        assert 'The username is already taken' in r.text
        r = register('', 'default')
        assert 'You have to enter a username' in r.text
        r = register('meh', '')
        assert 'You have to enter a password' in r.text
        r = register('meh', 'x', 'y')
        assert 'The two passwords do not match' in r.text
        r = register('meh', 'foo', email='broken')
        assert 'You have to enter a valid email address' in r.text

def test_login_logout():
    """Make sure logging in and logging out works"""
    r, http_session = register_and_login('user1', 'default')
    assert 'You were logged in' in r.text
    r = logout(http_session)
    assert 'You were logged out' in r.text
    r, _ = login('user1', 'wrongpassword')
    assert 'Invalid password' in r.text
    r, _ = login('user2', 'wrongpassword')
    assert 'Invalid username' in r.text

def test_message_recording():
    """Check if adding messages works"""
    _, http_session = register_and_login('foo', 'default')
    add_message(http_session, 'test message 1')
    add_message(http_session, '<test message 2>')
    r = requests.get(f'{BASE_URL}/')
    assert 'test message 1' in r.text
    assert '&lt;test message 2&gt;' in r.text

def test_timelines():
    """Make sure that timelines work"""
    _, http_session = register_and_login('foo', 'default')
    add_message(http_session, 'the message by foo')
    logout(http_session)
    _, http_session = register_and_login('bar', 'default')
    add_message(http_session, 'the message by bar')
    r = http_session.get(f'{BASE_URL}/public')
    assert 'the message by foo' in r.text
    assert 'the message by bar' in r.text

    # bar's timeline should just show bar's message
    r = http_session.get(f'{BASE_URL}/')
    assert 'the message by foo' not in r.text
    assert 'the message by bar' in r.text

    # now let's follow foo
    r = http_session.get(f'{BASE_URL}/foo/follow', allow_redirects=True)
    assert 'You are now following foo' in r.text

    # we should now see foo's message
    r = http_session.get(f'{BASE_URL}/')
    assert 'the message by foo' in r.text
    assert 'the message by bar' in r.text

    # but on the user's page we only want the user's message
    r = http_session.get(f'{BASE_URL}/bar')
    assert 'the message by foo' not in r.text
    assert 'the message by bar' in r.text
    r = http_session.get(f'{BASE_URL}/foo')
    assert 'the message by foo' in r.text
    assert 'the message by bar' not in r.text

    # now unfollow and check if that worked
    r = http_session.get(f'{BASE_URL}/foo/unfollow', allow_redirects=True)
    assert 'You are no longer following foo' in r.text
    r = http_session.get(f'{BASE_URL}/')
    assert 'the message by foo' not in r.text
    assert 'the message by bar' in r.text