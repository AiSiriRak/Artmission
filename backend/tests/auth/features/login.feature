Feature: Login
  As a registered user
  I want to log in with my email and password
  So that I can obtain an access token and a session

  Background:
    Given the user has a registered account

  Scenario: log in with valid credentials
    When the user logs in with valid credentials
    Then the system authenticates the user
    And the system starts a new session for the user

  Scenario: log in with an incorrect password
    When the user logs in with an incorrect password
    Then the system does not authenticate the user
    And the system displays an appropriate error message

  Scenario: log in with an unknown email
    When the user logs in with an unknown email
    Then the system does not authenticate the user
    And the system displays an appropriate error message
