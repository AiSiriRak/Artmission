Feature: Confirm Deliverable
  As a customer
  I want to accept or reject a deliverable for my order
  So that an artist knows whether the work is approved or needs revision

  Scenario: a customer accepts their order's pending deliverable
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order with a pending deliverable
    When the customer accepts their last deliverable
    Then the system confirms the deliverable and marks the order as "SUCCESS"

  Scenario: a customer cannot confirm another customer's deliverable
    Given the user has a registered customer account
    And the user has logged in
    And another customer has an order with a pending deliverable
    When the user confirms another customer's deliverable
    Then the system hides the deliverable from the user

  Scenario: a customer cannot confirm a deliverable through a different order
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order with a pending deliverable
    And the user has another order
    When the user confirms their last deliverable through another order
    Then the system hides the deliverable from the user
