Feature: Refresh
  As a logged-in user
  I want to exchange my refresh token for a new access token
  So that my session can outlive a short-lived access token

  Background:
    Given the user has a registered account
    And the user has logged in

  Scenario: refresh with a valid session
    When the user refreshes the session
    Then the system renews the session for the user

  Scenario: refresh rotates the refresh token
    When the user refreshes the session
    And the user refreshes the session again using the previous refresh token
    Then the system rejects the request

  Scenario: refresh without a refresh token
    When the user refreshes the session without a refresh token
    Then the system rejects the request

  Scenario: refresh with an invalid refresh token
    When the user refreshes the session with an invalid refresh token
    Then the system rejects the request
