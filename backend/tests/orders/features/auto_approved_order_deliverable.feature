Feature: Order deliverables are auto-approved

  Background:
    Given the user has a registered customer account
    And the user has logged in

  Scenario: the system auto-approves a deliverable after the review window
    Given the user has an order with a pending deliverable
    And the pending deliverable was submitted more than 3 days ago
    When the system auto-approves expired deliverables
    Then the system shows the deliverable as approved and the order as "SUCCESS"

  Scenario: the system does not auto-approve a deliverable when its order is not in progress
    Given the user has an order with status "PENDING" and a pending deliverable
    And the pending deliverable was submitted more than 3 days ago
    When the system auto-approves expired deliverables
    Then the system leaves the deliverable pending and the order unchanged
