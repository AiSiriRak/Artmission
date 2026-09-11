Feature: Artist artwork portfolio
  As an artist
  I want to create and delete my portfolio artwork
  So that customers can see my commission offerings

  Scenario: create artist-owned artwork with free labels and samples
    Given an artist is registered and logged in
    When the artist creates artwork with samples
    Then the artwork is stored with its category, styles, and ordered samples

  Scenario: delete artist-owned artwork
    Given an artist is registered and logged in
    And the artist has created artwork with samples
    When the artist deletes their artwork
    Then the artwork and its relations are deleted

  Scenario: hide artwork ownership from another artist
    Given an artist is registered and logged in
    And another artist has created artwork with samples
    When the artist deletes the other artist's artwork
    Then the system reports the artwork was not found and keeps it

  Scenario: require an authenticated artist to create artwork
    When an unauthenticated caller creates artwork with samples
    Then the system requires the caller to log in

  Scenario: prevent a customer from creating artwork
    Given a customer is registered and logged in
    When the customer creates artwork with samples
    Then the system denies artwork creation
