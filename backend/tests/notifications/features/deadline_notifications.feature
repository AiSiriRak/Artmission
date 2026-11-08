Feature: Deadline extension notifications
  Background:
    Given an artist and a customer are registered and logged in
    And an order with status "IN_PROCESS" exists

  Scenario: Requesting an extension notifies the customer
    When the artist requests a deadline extension
    Then the customer receives a "DEADLINE_EXTENSION_REQUESTED" notification
    And the notification references that extension and its proposed deadline

  Scenario Outline: The customer decision notifies the artist
    Given a deadline extension is awaiting a response
    When the customer <decision> the deadline extension
    Then the artist receives a "<type>" notification
    Examples:
      | decision | type                        |
      | approves | DEADLINE_EXTENSION_APPROVED  |
      | rejects  | DEADLINE_EXTENSION_REJECTED  |

  Scenario: Repeated extensions retain each round's deadline
    When the artist and customer approve two extension rounds
    Then both approval notifications retain their own deadline
