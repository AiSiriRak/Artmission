Feature: Logout
  As a logged-in user
  I want to log out
  So that my session can no longer be used

  Background:
    Given the user has a registered account
    And the user has logged in

  Scenario: log out while authenticated
    When the user logs out
    Then the system terminates the current session

  Scenario: logging out invalidates the session
    When the user logs out
    Then the system terminates the current session
    When the user refreshes the session using the previous refresh token
    Then the system rejects the request

  Scenario: logging out twice rejects the second attempt with the same access token
    When the user logs out
    Then the system terminates the current session
    When the user logs out
    Then the system does not terminate a session
    And the system displays an appropriate error message

  Scenario: log out without being authenticated
    When the user logs out without an access token
    Then the system does not terminate a session
    And the system displays an appropriate error message

  Scenario: log out with an invalid access token
    When the user logs out with an invalid access token
    Then the system does not terminate a session
    And the system displays an appropriate error message
