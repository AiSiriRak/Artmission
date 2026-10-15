Feature: Delay requests expire automatically

  Background:
    Given the user has a registered customer account
    And the user has logged in

  Scenario: the system auto-rejects a delay request after the response window
    Given the user has an order with status "IN_PROCESS" and a pending delay request
    And the pending delay request was created more than 2 days ago
    When the system auto-expires delay requests
    Then the system shows the delay request as rejected

  Scenario: the system does not auto-reject a delay request within the response window
    Given the user has an order with status "IN_PROCESS" and a pending delay request
    When the system auto-expires delay requests
    Then the system leaves the delay request pending
