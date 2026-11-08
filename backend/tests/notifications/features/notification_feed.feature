Feature: Personal notification feed
  Background:
    Given an artist and a customer are registered and logged in

  Scenario: Empty feeds return an array
    When the customer requests their notifications
    Then the feed contains 0 items and total 0

  Scenario: Notifications can only be marked read by their recipient
    Given an order with status "PENDING" exists
    When the customer marks the artist notification read
    Then the response status is 404

  Scenario: Marking read is idempotent and updates the unread count
    Given an order with status "PENDING" exists
    Then the artist has 1 unread notifications
    When the artist marks their notification read twice
    Then the artist has 0 unread notifications

  Scenario: Feed pagination preserves total and recipient isolation
    Given the customer has created 3 orders
    When the artist requests notifications with limit 1 and offset 1
    Then the feed contains 1 items and total 3
    And the customer has no notifications

  Scenario Outline: Authentication is required
    When a visitor requests "<path>" using "<method>"
    Then the response status is 401
    Examples:
      | method | path                                                        |
      | GET    | /notifications                                              |
      | GET    | /notifications/unread-count                                 |
      | PATCH  | /notifications/00000000-0000-0000-0000-000000000001/read        |
