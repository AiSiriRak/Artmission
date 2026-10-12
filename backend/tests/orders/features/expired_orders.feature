Feature: Expired orders are cancelled automatically

  Background:
    Given the user has a registered customer account
    And the user has logged in

  Scenario Outline: The system cancels an order that has expired
    Given the user has an order with status "<status>" created <created> and a deadline that <deadline>
    When the system cancels expired orders
    Then the system shows the order with status "<outcome>"

    Examples:
      | status     | created              | deadline       | outcome  |
      | PENDING    | just now             | has passed     | CANCEL   |
      | PENDING    | more than 3 days ago | is still ahead | CANCEL   |
      | NOT_PAID   | just now             | has passed     | CANCEL   |
      | IN_PROCESS | just now             | has passed     | CANCEL   |
      | SUCCESS    | just now             | has passed     | SUCCESS  |
      | CANCEL     | just now             | has passed     | CANCEL   |
      | PENDING    | just now             | is still ahead | PENDING  |
      | NOT_PAID   | just now             | is still ahead | NOT_PAID |

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
