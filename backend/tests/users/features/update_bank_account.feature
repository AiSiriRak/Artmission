Feature: Update bank account
  As an authenticated user
  I want to update my payout account
  So that future payments use my current details

  Scenario: update bank account with valid details
    Given the user has a registered account
    And the user has logged in
    When the user updates their bank account with valid details
    Then the system saves the updated bank account details

  Scenario: create a missing bank account through update
    Given the user has a registered account
    And the user has logged in
    And the user has no saved bank account
    When the user updates their bank account with valid details
    Then the system saves the updated bank account details

  Scenario: update bank account without logging in
    When the user updates a bank account without logging in
    Then the system requires the user to log in

  Scenario: update bank account with blank details
    Given the user has a registered account
    And the user has logged in
    When the user updates their bank account with a blank bank name
    Then the system rejects the bank account update
