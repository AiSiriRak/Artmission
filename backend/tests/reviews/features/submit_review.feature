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

    Examples:
      | body                                   | rating | comment     |
      | {"rating":1,"comment":"  Great work!  "} | 1      | Great work! |
      | {"rating":5}                           | 5      | NULL        |

  Scenario Outline: reject a missing or invalid rating
    Given a commission in status "SUCCESS"
    When the customer submits this review:
      """
      <body>
      """
    Then the review is rejected as "invalid request"

    Examples:
      | body                |
      | {"comment":"Hello"} |
      | {"rating":6}        |

  Scenario: reject a review before a commission succeeds
    Given a commission in status "IN_PROCESS"
    When the customer submits this review:
      """
      {"rating":5}
      """
    Then the review is rejected as "unfinished commission"

  Scenario Outline: restrict who can submit a review
    Given a commission in status "SUCCESS"
    When "<caller>" submits a review
    Then the review is rejected as "<reason>"

    Examples:
      | caller           | reason    |
      | another customer | forbidden |
      | the artist       | forbidden |

  Scenario: reject a second review for the same commission
    Given a commission in status "SUCCESS"
    And the customer has already reviewed the commission
    When the customer submits this review:
      """
      {"rating":1,"comment":"Replace the original"}
      """
    Then the review is rejected as "duplicate review"
    And the original review remains unchanged

  Scenario: calculate the artist average from all submitted reviews
    Given a commission in status "SUCCESS"
    When the customer reviews three completed commissions with ratings 5, 4, and 4
    Then the artist average score is 4.3
