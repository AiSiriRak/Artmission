Feature: Account information
  As an authenticated user
  I want to view and update my account information
  So that my username and password stay current

  Scenario: view account information
    Given the user has a registered account
    And the user has logged in
    When the user requests their account information
    Then the system returns their safe account information

  Scenario: view account information without logging in
    When the user requests account information without logging in
    Then the system requires the user to log in

  Scenario: update username
    Given the user has a registered account
    And the user has logged in
    When the user updates their username
    Then the system saves and returns the new username

  Scenario: change password
    Given the user has a registered account
    And the user has logged in
    When the user changes their password with the correct old password
    Then the new password works and the current session remains valid

  Scenario: reject an incorrect old password
    Given the user has a registered account
    And the user has logged in
    When the user changes their password with an incorrect old password
    Then the account information update is unauthorized and no changes are saved

  Scenario: reject an incomplete password change
    Given the user has a registered account
    And the user has logged in
    When the user sends only a new password
    Then the system rejects the account information update
