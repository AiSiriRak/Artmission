Feature: Register
  As a new user
  I want to create an account
  So that I can log in and use the platform

  Scenario: register with valid details
    Given the user does not have an account
    When the user registers with a valid username, password, phone number, and email
    Then the system creates the account

  Scenario Outline: register with invalid details
    Given the user does not have an account
    When the user registers with <field> "<value>"
    Then the system does not create the account
    And the system displays an appropriate error message

    Examples:
      | field        | value        |
      | username     | ab           |
      | email        | not-an-email |
      | password     | short1       |
      | phone_number |              |

  Scenario: register with a username that is already in use
    Given the user has a registered account
    When the user registers reusing that account's username
    Then the system does not create the account
    And the system displays an appropriate error message

  Scenario: register with an email that is already in use
    Given the user has a registered account
    When the user registers reusing that account's email
    Then the system does not create the account
    And the system displays an appropriate error message

  Scenario: register as an artist with a profile description
    Given the user does not have an account
    When the user registers as an artist with a profile description
    Then the system creates the account

  Scenario: register as an artist without a profile description
    Given the user does not have an account
    When the user registers as an artist without a profile description
    Then the system does not create the account
    And the system displays an appropriate error message
