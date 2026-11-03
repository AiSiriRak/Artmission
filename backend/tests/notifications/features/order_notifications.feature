Feature: Order notifications
  As an artist or customer
  I want updates about my commission
  So that I can respond to order changes

  Background:
    Given an artist and a customer are registered and logged in

  Scenario: A new order notifies its artist
    When the customer creates an order
    Then the artist receives an "ORDER_CREATED" notification
    And the customer has no notifications

  Scenario Outline: An artist decision notifies the customer
    Given an order with status "PENDING" exists
    When the artist <decision> the order
    Then the customer receives an "<type>" notification
    Examples:
      | decision | type           |
      | accepts  | ORDER_ACCEPTED |
      | declines | ORDER_DECLINED |

  Scenario: Payment bypass notifies the artist only once
    Given an order with status "NOT_PAID" exists
    When the customer bypasses payment
    Then the artist receives a "PAYMENT_SECURED" notification
    When the customer repeats payment bypass
    Then the request fails without adding notifications

  Scenario: A draft links its source version
    Given an order with status "IN_PROCESS" exists
    When the artist submits a draft
    Then the customer receives a "DRAFT_SUBMITTED" notification
    And the notification references the submitted draft and its version

  Scenario Outline: A customer decision notifies the artist
    Given a draft is awaiting review
    When the customer <decision> the draft
    Then the artist receives an "<type>" notification
    Examples:
      | decision | type               |
      | accepts  | ORDER_COMPLETED    |
      | rejects  | REVISION_REQUESTED |

  Scenario Outline: Cancellation notifies the other participant
    Given an order with status "IN_PROCESS" exists
    When the <actor> cancels the order
    Then the <recipient> receives an "ORDER_CANCELED" notification
    Examples:
      | actor    | recipient |
      | artist   | customer  |
      | customer | artist    |

  Scenario: The last allowed draft makes final artwork ready
    Given two drafts have been rejected
    When the artist submits a draft
    Then the customer receives a "FINAL_ARTWORK_READY" notification
