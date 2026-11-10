Feature: Order review access
  As an order participant
  I want to read the order review
  So that I can see the customer's feedback

  Scenario Outline: restrict access to an existing order review
    Given a commission in status "SUCCESS"
    And the customer has already reviewed the commission
    And the customer has a profile image
    When "<caller>" requests the order review
    Then the order review response has status <status>
    And the order review includes the customer's public profile image URL

    Examples:
      | caller           | status |
      | the customer     | 200    |
      | the artist       | 200    |
      | another customer | 403    |

  Scenario: an order has no review
    Given a commission in status "SUCCESS"
    When "the customer" requests the order review
    Then the order review response has status 404