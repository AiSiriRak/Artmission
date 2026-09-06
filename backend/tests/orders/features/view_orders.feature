Feature: View Orders
  As a customer or an artist
  I want to view the orders where I'm a participant, filtered, sorted, and paginated
  So that I can track my commissions

  Scenario: view orders with existing orders
    Given the user has a registered customer account
    And the user has logged in
    And the user has one or more orders
    When the user views their orders
    Then the system shows all of the user's orders with their current status

  Scenario: a customer's orders exclude another customer's orders
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order
    And another customer has an order
    When the user views their orders
    Then the system does not show any other user's orders

  Scenario: an artist views their orders
    Given the user has a registered artist account
    And the user has logged in
    And the user has one or more orders
    When the user views their orders
    Then the system shows all of the user's orders with their current status

  Scenario: an artist's orders exclude another artist's orders
    Given the user has a registered artist account
    And the user has logged in
    And the user has an order
    And another artist has an order
    When the user views their orders
    Then the system does not show any other user's orders

  Scenario: view orders with no orders
    Given the user has a registered customer account
    And the user has logged in
    When the user views their orders
    Then the system shows an empty order list

  Scenario: view orders while unauthenticated
    Given the user has a registered customer account
    When the user views their orders without logging in
    Then the system requires the user to log in

  Scenario: filtering orders by status
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order with status "PENDING"
    And the user has an order with status "SUCCESS"
    When the user views their orders filtered by status "SUCCESS"
    Then the system shows only orders with status "SUCCESS"

  Scenario: paginating through orders with a small limit
    Given the user has a registered customer account
    And the user has logged in
    And the user has 5 orders
    When the user pages through all of their orders using a limit of 2
    Then the system returns every seeded order exactly once

  Scenario: the system reports the total number of matching orders
    Given the user has a registered customer account
    And the user has logged in
    And the user has 5 orders
    When the user views their orders with a limit of 2 and an offset of 0
    Then the system reports a total of 5 orders

  Scenario: revisiting an earlier offset returns the same page
    Given the user has a registered customer account
    And the user has logged in
    And the user has 4 orders
    When the user views their orders with a limit of 2 and an offset of 0
    And the user remembers the current page as the first page
    And the user views their orders with a limit of 2 and an offset of 2
    And the user views their orders with a limit of 2 and an offset of 0
    Then the system returns the first page of orders again

  Scenario: a negative offset is rejected
    Given the user has a registered customer account
    And the user has logged in
    When the user views their orders with an invalid offset
    Then the system rejects the request due to an invalid offset

  Scenario: an invalid status is rejected
    Given the user has a registered customer account
    And the user has logged in
    When the user views their orders filtered by status "NOT_A_STATUS"
    Then the system rejects the request due to an invalid status

  Scenario: sorting by deadline places orders without a deadline last, ascending
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order with deadline "2026-06-01T00:00:00Z"
    And the user has an order with deadline "2026-01-01T00:00:00Z"
    And the user has an order with no deadline
    When the user views their orders sorted by deadline in "asc" order
    Then the system returns orders sorted by deadline in "asc" order

  Scenario: sorting by deadline places orders without a deadline first, descending
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order with deadline "2026-06-01T00:00:00Z"
    And the user has an order with deadline "2026-01-01T00:00:00Z"
    And the user has an order with no deadline
    When the user views their orders sorted by deadline in "desc" order
    Then the system returns orders sorted by deadline in "desc" order
