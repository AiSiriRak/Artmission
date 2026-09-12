Feature: Artist artwork portfolio
  As an artist
  I want to create, update, and delete my portfolio artwork
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

  Scenario: update artist-owned artwork and selected samples
    Given an artist is registered and logged in
    And the artist has created artwork with samples
    When the artist updates their artwork details and selected samples
    Then the artwork contains the updated values and retained samples

  Scenario: clear artwork styles and samples
    Given an artist is registered and logged in
    And the artist has created artwork with samples
    When the artist replaces their artwork with empty styles and samples
    Then the artwork has no styles or samples

  Scenario: reject deleting a sample outside the artwork
    Given an artist is registered and logged in
    And the artist has created artwork with samples
    When the artist updates the artwork with an unknown deleted sample URL
    Then the system rejects the sample deletion and leaves the artwork unchanged

  Scenario: hide artwork ownership during update
    Given an artist is registered and logged in
    And another artist has created artwork with samples
    When the artist updates the other artist's artwork
    Then the system reports the artwork was not found and leaves it unchanged

  Scenario: report a missing artwork during update
    Given an artist is registered and logged in
    When the artist updates a missing artwork
    Then the system reports the artwork was not found

  Scenario: require authentication to update artwork
    Given an artist is registered and logged in
    And the artist has created artwork with samples
    When an unauthenticated caller updates the artwork
    Then the system requires the caller to log in

  Scenario: prevent a customer from updating artwork
    Given an artist is registered and logged in
    And the artist has created artwork with samples
    And a customer is registered and logged in
    When the customer updates the artwork
    Then the system denies artwork updates
