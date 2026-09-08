Feature: Delete account
  As an authenticated user
  I want to delete my account
  So that it can no longer be used to access the platform

  Scenario: delete an account without active orders
    Given the user has a registered account
    And the user has logged in
    When the user deletes their account
    Then the account is soft-deleted and all private access data is removed

  Scenario: reject account deletion without authentication
    When a user deletes an account without logging in
    Then the system requires the user to log in

  Scenario: prevent the customer from deleting during an active order
    Given a customer and artist have an active order
    And the customer has logged in
    When the user deletes their account
    Then the system rejects account deletion because the order is active

  Scenario: prevent the artist from deleting during an active order
    Given a customer and artist have an active order
    And the artist has logged in
    When the user deletes their account
    Then the system rejects account deletion because the order is active
