Feature: Reviews by artwork
  As a customer
  I want to read reviews for a selected artwork
  So that I can see feedback about that artwork

  Scenario: an artwork has no reviews
    Given a commission in status "SUCCESS"
    When the API receives a "artwork" review request with limit 20 and offset 0
    Then the review page contains 0 reviews with total 0

  Scenario: exclude reviews of another artwork by the same artist
    Given a commission in status "SUCCESS"
    And the artist has reviews whose order dates differ from their review dates
    And another artwork by the same artist has a review
    And the customer has a profile image
    When the API receives a "artwork" review request with limit 20 and offset 0
    Then the review page contains 3 reviews with total 3
    And the reviews contain the customer and order names
    And the review details include the customer's public profile image URL
    And the review page matches the expected order from offset 0
