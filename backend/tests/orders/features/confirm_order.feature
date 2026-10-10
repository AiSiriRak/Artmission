Feature: Confirm Order
  As an artist
  I want to accept or reject a pending order
  So that I can decide whether to take on the commission

  Scenario: an artist accepts a pending order
    Given the user has a registered artist account
    And the user has logged in
    And the user has an order
    When the artist accepts their last order
    Then the system confirms the order with status "NOT_PAID"

  Scenario: an artist rejects a pending order
    Given the user has a registered artist account
    And the user has logged in
    And the user has an order
    When the artist rejects their last order
    Then the system confirms the order with status "CANCEL"

  Scenario: an artist cannot confirm an order that is no longer pending
    Given the user has a registered artist account
    And the user has logged in
    And the user has an order with status "SUCCESS"
    When the artist accepts their last order
    Then the system rejects the confirmation because the order status does not allow it

  Scenario: a customer cannot confirm an order
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order
    When the user attempts to accept their last order
    Then the system forbids the order confirmation
