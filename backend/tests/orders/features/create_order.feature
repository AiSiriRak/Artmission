Feature: Get Order
    As a customer
    I want to create an order
    So that the artist can accept the order

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
