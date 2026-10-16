# for sprint 2
Feature: Bypass Order Payment
  As a customer
  I want to bypass payment during Sprint 2 testing
  So that I can move an accepted order into production

  Scenario: a customer bypasses payment for their unpaid order
    Given the user has a registered customer account
    And the user has logged in
    And the user has an order with status "NOT_PAID"
    When the user bypasses payment for their last order
    Then the system confirms payment bypass and moves the order to IN_PROCESS

  Scenario: an artist cannot bypass payment
    Given the user has a registered artist account
    And the user has logged in
    And the user has an order with status "NOT_PAID"
    When the user bypasses payment for their last order
    Then the system forbids payment bypass

  Scenario: an unauthenticated user cannot bypass payment
    Given the user has a registered customer account
    And the user has an order with status "NOT_PAID"
    When the user attempts to bypass payment without logging in
    Then the system requires the user to log in
