Feature: Review pagination
  As a customer
  I want to see the first 20 reviews and load more by clicking See more
  So that I can read additional reviews without leaving the page

  # A limit of 2 keeps the API pagination fixture small; the default is 20.
  Scenario Outline: retrieve a review page with the full total
    Given a commission in status "SUCCESS"
    And the artist has reviews whose order dates differ from their review dates
    When the API receives a "<resource>" review request with limit 2 and offset <offset>
    Then the review page contains <count> reviews with total 3
    And the review page matches the expected order from offset <offset>

    Examples:
      | resource | offset | count |
      | artist   | 0      | 2     |
      | artist   | 2      | 1     |
      | artist   | 3      | 0     |
      | artwork  | 0      | 2     |
      | artwork  | 2      | 1     |
      | artwork  | 3      | 0     |
