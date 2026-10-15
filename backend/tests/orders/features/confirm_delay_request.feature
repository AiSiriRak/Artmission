Feature: Confirm Order Delay Request
  As a customer
  I want to approve or reject an artist's deadline extension request
  So that the order deadline reflects my decision

  Scenario: a customer approves a pending delay request
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order with status "IN_PROCESS" and a pending delay request
    When the customer approves the pending delay request for their last order
    Then the system confirms the delay request as "APPROVED"

  Scenario: a customer rejects a pending delay request
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order with status "IN_PROCESS" and a pending delay request
    When the customer rejects the pending delay request for their last order
    Then the system confirms the delay request as "REJECTED"

  Scenario: an artist cannot decide a delay request
    Given the user has a registered artist account
    And the user has logged in
    And the user has an order with status "IN_PROCESS" and a pending delay request
    When the artist attempts to approve the pending delay request for their last order
    Then the system forbids the delay request decision

  Scenario: a customer cannot decide a delay request without logging in
    Given the user has a registered customer account
    And the user has an order with status "IN_PROCESS" and a pending delay request
    When the user attempts to approve the pending delay request without logging in
    Then the system requires the user to log in

  Scenario: a customer cannot decide a delay request when the order is not in progress
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order with status "PENDING" and a pending delay request
    When the customer approves the pending delay request for their last order
    Then the system rejects the delay request decision because the order status does not allow it
