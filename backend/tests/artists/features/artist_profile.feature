Feature: Artist profile
  As a visitor or artist
  I want to retrieve or maintain artist profile information
  So that commission preferences are accurate and publicly visible

  Scenario: retrieve a newly registered artist profile publicly
    Given the artist has a registered account
    When a visitor requests the artist profile
    Then the system returns the initial public artist profile

  Scenario: retrieve an artist profile with no description
    Given the artist has a registered account without a description
    When a visitor requests the artist profile
    Then the system returns a null artist description

  Scenario: retrieve artwork categories and review score
    Given the artist has a registered account
    And the artist has samples in multiple categories
    And the artist has a review score
    When a visitor requests the artist profile
    Then the system returns distinct artwork categories and the review score

  Scenario: update an artist profile
    Given the artist has a registered account
    And reference styles are available
    And the artist has logged in
    When the artist updates their profile with valid details
    Then the system saves and returns the complete artist profile

  Scenario: clear every selected style
    Given the artist has a configured profile
    When the artist clears their selected styles
    Then the system returns an empty style selection

  Scenario: reject an invalid price range atomically
    Given the artist has a configured profile
    When the artist updates their profile with an invalid price range
    Then the system rejects the artist update without changing the profile

  Scenario: reject an unknown style atomically
    Given the artist has a configured profile
    When the artist updates their profile with an unknown style
    Then the system rejects the artist update without changing the profile

  Scenario: require authentication to update a profile
    Given the artist has a registered account
    When someone updates the artist profile without logging in
    Then the system requires the user to log in

  Scenario: prevent a customer from updating an artist profile
    Given the customer has logged in
    When the customer updates the artist profile
    Then the system denies the artist profile update

  Scenario: retrieve a missing artist profile
    When a visitor requests an artist profile that does not exist
    Then the system reports that the artist profile was not found

  Scenario: hide a deleted artist profile
    Given the artist has a registered account
    And the artist account has been deleted
    When a visitor requests the artist profile
    Then the system reports that the artist profile was not found
