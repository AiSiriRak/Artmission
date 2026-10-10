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
