Feature: Artist review history
  As a customer or artist
  I want to read an artist's reviews newest first
  So that I can see the latest feedback

  Scenario Outline: retrieve reviews by review date rather than order date
    Given a commission in status "SUCCESS"
    And the artist has reviews whose order dates differ from their review dates
    When "<caller>" requests the artist reviews
    Then the artist reviews contain the customer and purchased artwork names
    And the artist reviews are returned newest first
    And reviews from other artists are excluded

    Examples:
      | caller       |
      | the customer |
      | the artist   |

  Scenario: use a stable order when review timestamps are equal
    Given a commission in status "SUCCESS"
    And the artist has reviews with identical timestamps
    When "the customer" requests the artist reviews
    Then the artist reviews are returned newest first

  Scenario: retrieve reviews after the source artwork is deleted
    Given a commission in status "SUCCESS"
    And the customer has already reviewed the commission
    And the reviewed artwork has been deleted
    When "the customer" requests the artist reviews
    Then the artist reviews contain the customer and purchased artwork names

  Scenario: show the empty state for an artist with no reviews
    Given a commission in status "SUCCESS"
    When "the customer" requests the artist reviews
    Then the system returns an empty review list and no average score

  Scenario: reject an unauthenticated review list request
    Given a commission in status "SUCCESS"
    When "a visitor" requests the artist reviews
    Then the artist review list requires authentication
