Feature: Escrow settlement and simulated wallet credit

  Scenario: Customer cancels a paid order with an odd satang
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order with status "IN_PROCESS"
    And the last order has a "HELD" payment of 10001 satang
    When the user cancels their last order
    Then the system cancels the order
    And the last order summary shows "CUSTOMER_CANCELLED" with refund 5000 and artist payout 5001 satang
    And the user's wallet shows 5000 satang and 1 incoming transactions

  Scenario: Artist cancellation allocates a full refund
    Given the user has a registered artist account
    And the user has logged in
    And the user has an order with status "IN_PROCESS"
    And the last order has a "HELD" payment of 10001 satang
    When the user cancels their last order
    Then the system cancels the order
    And the last order summary shows "ARTIST_CANCELLED" with refund 10001 and artist payout 0 satang
    And the user's wallet shows 0 satang and 0 incoming transactions

  Scenario: Unpaid cancellation has no allocation
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order with status "NOT_PAID"
    And the last order has a "PENDING" payment of 10001 satang
    When the user cancels their last order
    Then the system cancels the order
    And the last order summary shows "UNPAID" with refund 0 and artist payout 0 satang
    And the user's wallet shows 0 satang and 0 incoming transactions

  Scenario: Deadline expiration allocates a full refund
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order with status "IN_PROCESS" created just now and a deadline that has passed
    And the last order has a "HELD" payment of 10001 satang
    When the system cancels expired orders
    Then the last order summary shows "DEADLINE_EXPIRED" with refund 10001 and artist payout 0 satang
    And the user's wallet shows 10001 satang and 1 incoming transactions

  Scenario: Deliverable approval allocates the artist payout
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order with a pending deliverable
    And the last order has a "HELD" payment of 10001 satang
    When the customer accepts their last deliverable
    Then the system confirms the deliverable and marks the order as "SUCCESS"
    And the last order summary shows "COMPLETED" with refund 0 and artist payout 10001 satang

  Scenario: Automatic approval allocates the artist payout
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order with a pending deliverable
    And the last order has a "HELD" payment of 10001 satang
    And the pending deliverable was submitted more than 3 days ago
    When the system auto-approves expired deliverables
    Then the last order summary shows "COMPLETED" with refund 0 and artist payout 10001 satang

  Scenario: New users have an empty wallet and history
    Given the user has a registered customer account
    And the user has logged in
    Then the user's wallet shows 0 satang and 0 incoming transactions
