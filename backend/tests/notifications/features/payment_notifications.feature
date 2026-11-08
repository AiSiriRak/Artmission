Feature: QR and slip payment notifications
  Background:
    Given an artist and a customer are registered and logged in
    And an order with status "NOT_PAID" exists

  Scenario: Verified payment notifies the artist only once
    When the customer pays with a verified slip
    Then the artist receives a "PAYMENT_SECURED" notification
    When the customer submits the verified slip again
    Then the request fails without adding notifications

  Scenario: An invalid slip does not notify the artist
    When the customer uploads an invalid slip
    Then the order remains unpaid without new notifications

  Scenario: Payment expiry notifies both participants only once
    Given the payment QR expires
    When the payment expiry process runs
    Then the artist receives an "ORDER_CANCELED" notification
    And the customer receives an "ORDER_CANCELED" notification
    When the payment expiry process runs again
    Then no extra notifications are created
