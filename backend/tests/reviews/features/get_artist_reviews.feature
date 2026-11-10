Feature: Artist review history
  As a customer or artist
  I want to read an artist's reviews newest first
  So that I can see the latest feedback

  Scenario Outline: retrieve reviews by review date rather than order date
    Given a commission in status "SUCCESS"
    And the artist has reviews whose order dates differ from their review dates
    And the customer has a profile image
    When "<caller>" requests the artist reviews
    Then the reviews contain the customer and order names
    And the review details include the customer's public profile image URL
    And the artist reviews include artwork names from order snapshots
    And the artist reviews are returned newest first
    And reviews from other artists are excluded

    Examples:
      | caller       |
      | the customer |
      | the artist   |

  Scenario: show the empty state for an artist with no reviews
    Given a commission in status "SUCCESS"
    When "the customer" requests the artist reviews
    Then the system returns an empty review list and no average score
