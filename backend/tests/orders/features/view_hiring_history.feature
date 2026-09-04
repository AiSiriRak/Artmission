Feature: View Hiring History
  As a customer
  I want to view my hiring history and order statuses
  So that I can track my commissions

  Background:
    Given the user has a registered customer account

  Scenario: view hiring history with existing orders
    Given the user has logged in
    And the user has one or more orders
    When the user views their hiring history
    Then the system shows all of the user's orders with their current status

  Scenario: hiring history excludes other customers' orders
    Given the user has logged in
    And the user has an order
    And another customer has an order
    When the user views their hiring history
    Then the system does not show any other customer's orders

  Scenario: view hiring history with no orders
    Given the user has logged in
    When the user views their hiring history
    Then the system shows an empty hiring history

  Scenario: view hiring history while unauthenticated
    When the user views their hiring history without logging in
    Then the system requires the user to log in

  Scenario: an artist cannot view hiring history
    Given the user has a registered artist account
    And the user has logged in
    When the user views their hiring history
    Then the system forbids the request
