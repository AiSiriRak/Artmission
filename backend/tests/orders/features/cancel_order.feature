Feature: Cancel Order
  As a customer or artist
  I want to cancel an order that I participate in
  So that the order can be stopped when it should not continue

  Scenario: a customer cancels their pending order
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order
    When the user cancels their last order
    Then the system cancels the order

  Scenario: an artist cancels their pending order
    Given the user has a registered artist account
    And the user has logged in
    And the user has an order
    When the user cancels their last order
    Then the system cancels the order

  Scenario: a customer cancels an unpaid order
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order with status "NOT_PAID"
    When the user cancels their last order
    Then the system cancels the order

  Scenario: an artist cancels an in-process order
    Given the user has a registered artist account
    And the user has logged in
    And the user has an order with status "IN_PROCESS"
    When the user cancels their last order
    Then the system cancels the order

  Scenario: a customer cannot cancel a completed order
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order with status "SUCCESS"
    When the user cancels their last order
    Then the system rejects cancellation because the order status does not allow it

  Scenario: a customer cannot cancel an already cancelled order
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order with status "CANCEL"
    When the user cancels their last order
    Then the system rejects cancellation because the order status does not allow it

  Scenario: a customer cannot cancel an order that belongs to another customer
    Given the user has a registered customer account
    And the user has logged in
    And another customer has an order
    When the user cancels another user's order
    Then the system rejects cancellation because the order is not owned

  Scenario: a user cannot cancel an order without logging in
    Given the user has a registered customer account
    And the user has an order
    When the user cancels their last order without logging in
    Then the system requires the user to log in to cancel
