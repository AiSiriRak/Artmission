Feature: Create Order Delay Request
  As an artist
  I want to request more time to complete an order
  So that the customer can review a revised deadline

  Scenario: an artist requests a later deadline for an order in progress
    Given the user has a registered artist account
    And the user has logged in
    And the user has an order with status "IN_PROCESS"
    When the artist requests a deadline extension for their last order
    Then the system creates the delay request successfully

  Scenario: a customer cannot request a deadline extension
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order with status "IN_PROCESS"
    When the artist requests a deadline extension for their last order
    Then the system forbids the delay request

  Scenario: an artist cannot request an extension for an order not in progress
    Given the user has a registered artist account
    And the user has logged in
    And the user has an order with status "PENDING"
    When the artist requests a deadline extension for their last order
    Then the system rejects the delay request because the order status does not allow it

  Scenario: an artist cannot request a deadline that is not later
    Given the user has a registered artist account
    And the user has logged in
    And the user has an order with status "IN_PROCESS"
    When the artist requests an extension that is not later than the current deadline
    Then the system rejects the delay request because its deadline is invalid

  Scenario: an artist cannot create another request while one awaits a decision
    Given the user has a registered artist account
    And the user has logged in
    And the user has an order with status "IN_PROCESS" and a pending delay request
    When the artist requests a deadline extension for their last order
    Then the system rejects the delay request because one is already pending

  Scenario: an unauthenticated user cannot request a deadline extension
    Given the user has a registered artist account
    And the user has an order with status "IN_PROCESS"
    When the artist requests a deadline extension for their last order without logging in
    Then the system requires the user to log in
