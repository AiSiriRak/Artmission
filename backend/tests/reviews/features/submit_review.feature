Feature: Customer reviews of completed commissions
  As a customer
  I want to review an artist after a successful commission
  So that other customers can learn about my experience

  Scenario Outline: submit a valid rating with an optional comment
    Given a commission in status "SUCCESS"
    When the customer submits this review:
      """
      <body>
      """
    Then the review is saved with rating <rating> and comment "<comment>"
    And the artist average score is <average>

    Examples:
      | body                                          | rating | comment     | average |
      | {"rating":1,"comment":"  Great work!  "}     | 1      | Great work! | 1.0     |
      | {"rating":5}                                  | 5      | NULL        | 5.0     |
      | {"rating":3,"comment":null}                   | 3      | NULL        | 3.0     |
      | {"rating":4,"comment":""}                   | 4      | NULL        | 4.0     |
      | {"rating":5,"comment":"   "}                | 5      | NULL        | 5.0     |

  Scenario Outline: reject a missing or invalid rating
    Given a commission in status "SUCCESS"
    When the customer submits this review:
      """
      <body>
      """
    Then the review is rejected as "invalid request"

    Examples:
      | body                  |
      | {"comment":"Hello"} |
      | {"rating":null}       |
      | {"rating":0}          |
      | {"rating":-1}         |
      | {"rating":6}          |
      | {"rating":1.5}        |
      | {"rating":"5"}      |

  Scenario Outline: reject a review before a commission succeeds
    Given a commission in status "<status>"
    When the customer submits this review:
      """
      {"rating":5}
      """
    Then the review is rejected as "unfinished commission"

    Examples:
      | status     |
      | PENDING    |
      | NOT_PAID   |
      | IN_PROCESS |
      | CANCEL     |

  Scenario Outline: restrict who can submit a review
    Given a commission in status "SUCCESS"
    When "<caller>" submits a review
    Then the review is rejected as "<reason>"

    Examples:
      | caller           | reason          |
      | another customer | forbidden       |
      | the artist       | forbidden       |
      | a visitor        | unauthenticated |
      | an expired token | unauthenticated |

  Scenario: reject a review for a missing order
    Given a commission in status "SUCCESS"
    When the customer reviews an order that does not exist
    Then the review is rejected as "missing order"

  Scenario: reject a second review for the same commission
    Given a commission in status "SUCCESS"
    And the customer has already reviewed the commission
    When the customer submits this review:
      """
      {"rating":1,"comment":"Replace the original"}
      """
    Then the review is rejected as "duplicate review"
    And the original review remains unchanged

  Scenario: accept only one of two simultaneous reviews
    Given a commission in status "SUCCESS"
    When the customer submits the same review concurrently
    Then only one review is saved

  Scenario: calculate the artist average from all submitted reviews
    Given a commission in status "SUCCESS"
    When the customer reviews three completed commissions with ratings 5, 4, and 4
    Then the artist average score is 4.3
