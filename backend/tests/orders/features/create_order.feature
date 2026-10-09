Feature: Get Order
    As a customer
    I want to create an order
    So that the artist can accept the order

  Scenario: a customer views their order detail
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order
    When the user views their last order
    Then the system shows the order details for the user

  Scenario: an artist views their order detail
    Given the user has a registered artist account
    And the user has logged in
    And the user has an order
    When the user views their last order
    Then the system shows the order details for the user

  Scenario: an unauthenticated user cannot view order detail
    Given the user has a registered customer account
    And the user has an order
    When the user views their last order without logging in
    Then the system requires the user to log in

  Scenario: a customer cannot view another customer's order detail
    Given the user has a registered customer account
    And the user has logged in
    And another customer has an order
    When the user views another user's order
    Then the system hides the order from the user

  Scenario: an artist cannot view another artist's order detail
    Given the user has a registered artist account
    And the user has logged in
    And another artist has an order
    When the user views another user's order
    Then the system hides the order from the user

  Scenario: a customer successfully creates an order
    Given the user has a registered customer account
    And the user has logged in
    And an artwork exists for commission
    When the user submits a new order for the artwork
    Then the system creates the order successfully with status "PENDING"

  Scenario: an artist cannot create an order
    Given the user has a registered artist account
    And the user has logged in
    And an artwork exists for commission
    When the user submits a new order for the artwork
    Then the system rejects the request due to forbidden role

  Scenario: an unauthenticated user cannot create an order
    Given an artwork exists for commission
    When the user submits a new order without logging in
    Then the system requires the user to log in
