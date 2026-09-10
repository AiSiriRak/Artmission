Feature: Artist profile
  As a visitor or artist
  I want artist profiles to reflect portfolio work and customer reviews
  So that commission information is accurate and publicly visible

  Scenario: retrieve a newly registered artist profile publicly
    Given the artist has a registered account
    When a visitor requests the artist profile
    Then the system returns an empty initial artist profile

  Scenario: derive categories, styles, and prices from artworks
    Given the artist has a registered account
    And the artist has artworks in multiple categories and styles
    When a visitor requests the artist profile
    Then the system returns distinct artwork metadata and its price range

  Scenario: return paginated reviews and a whole-profile average
    Given the artist has a registered account
    And the artist has three customer reviews
    When a visitor requests two artist reviews
    Then the system returns the newest two reviews and the complete average

  Scenario: update only the artist description
    Given the artist has a registered account
    And the artist has logged in
    When the artist updates only their description
    Then the system trims and saves the artist description

  Scenario: clear the artist description
    Given the artist has a registered account
    And the artist has logged in
    When the artist updates their description with blank text
    Then the system returns a null artist description

  Scenario: upload, replace, and remove a public profile image
    Given the artist has a registered account
    And the artist has logged in
    When the artist uploads a profile image
    Then the profile image is publicly accessible
    When the artist replaces the profile image
    Then the replacement profile image is publicly accessible and has a new URL
    When the artist removes the profile image
    Then the artist profile has no profile image

  Scenario: reject invalid profile image content without changing the profile
    Given the artist has a registered account
    And the artist has logged in
    When the artist uploads invalid profile image content
    Then the system rejects the artist update without changing the profile

  Scenario: reject legacy editable fields
    Given the artist has a registered account
    And the artist has logged in
    When the artist tries to update an artwork-derived price
    Then the system rejects the unsupported artist field

  Scenario: reject an empty profile update
    Given the artist has a registered account
    And the artist has logged in
    When the artist submits no profile changes
    Then the system rejects the empty artist update

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
