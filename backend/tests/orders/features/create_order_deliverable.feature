Feature: Create Order Deliverable
  As an artist
  I want to submit a deliverable for an order in progress
  So that the customer can review the work

  Scenario: an artist submits a deliverable for an order in progress
    Given the user has a registered artist account
    And the user has logged in
    And the user has an order with status "IN_PROCESS"
    When the user submits a deliverable for their last order
    Then the system creates the deliverable successfully

  Scenario: a customer cannot submit a deliverable
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order with status "IN_PROCESS"
    When the user submits a deliverable for their last order
    Then the system forbids the deliverable submission

  Scenario: an artist cannot submit a deliverable for an order that is not in progress
    Given the user has a registered artist account
    And the user has logged in
    And the user has an order with status "PENDING"
    When the user submits a deliverable for their last order
    Then the system rejects the deliverable because the order status does not allow it

  Scenario: an unauthenticated user cannot submit a deliverable
    Given the user has a registered artist account
    And the user has an order with status "IN_PROCESS"
    When the user submits a deliverable for their last order without logging in
    Then the system requires the user to log in
