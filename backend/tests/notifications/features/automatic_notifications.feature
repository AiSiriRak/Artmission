Feature: Automatic notification processing
  Background:
    Given an artist and a customer are registered and logged in

  Scenario Outline: Automatic processing notifies both participants only once
    Given an expired "<record>" exists
    When the expiry process runs
    Then the artist receives an "<artist_type>" notification
    And the customer receives a "<customer_type>" notification
    When the expiry process runs again
    Then no extra notifications are created
    Examples:
      | record        | artist_type                 | customer_type               |
      | order         | ORDER_CANCELED              | ORDER_CANCELED              |
      | draft         | ORDER_COMPLETED             | FINAL_ARTWORK_READY         |
      | delay request | DEADLINE_EXTENSION_REJECTED | DEADLINE_EXTENSION_REJECTED  |

  Scenario: Notification failure rolls back an order change
    Given an order with status "PENDING" exists
    When notification storage fails during order acceptance
    Then the order remains pending without new notifications

  Scenario: Failure for the second recipient rolls back the entire worker batch
    Given an expired "order" exists
    When notification storage fails for the second worker recipient
    Then the order remains pending without new notifications

  Scenario: Recording a payment ledger entry alone does not notify
    Given an order with status "IN_PROCESS" exists
    When a payment ledger entry is recorded
    Then no extra notifications are created
