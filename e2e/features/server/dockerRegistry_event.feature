Feature: API Endpoint /dockerRegistry/event

  Scenario: Accepts an event triggered by the push of a manifest to the Docker Registry
    Given the Docker image "127.0.0.1:52854/golang:1.12.4" pushed
    And imagespy API started
    When sending the request "fixtures/requests/a.http" to "/dockerRegistry/event"
    And waiting for "500ms"
    And sending a "GET" request to "/v2/images/127.0.0.1:52854/golang:1.12.4"
    Then the API responds with status code "200"
    And the API responds with a body of "fixtures/results/c.json"

  Scenario: Ignores an event triggered by the pull of a manifest from the Docker Registry
    Given the Docker image "127.0.0.1:52854/golang:1.12.4" pushed
    And imagespy API started
    When sending the request "fixtures/requests/b.http" to "/dockerRegistry/event"
    And waiting for "500ms"
    And sending a "GET" request to "/v2/images/127.0.0.1:52854/golang:1.12.4"
    Then the API responds with status code "404"

  Scenario: Ignores an event triggered by the push of a layer to the Docker Registry
    Given the Docker image "127.0.0.1:52854/golang:1.12.4" pushed
    And imagespy API started
    When sending the request "fixtures/requests/c.http" to "/dockerRegistry/event"
    And waiting for "500ms"
    And sending a "GET" request to "/v2/images/127.0.0.1:52854/golang:1.12.4"
    Then the API responds with status code "404"
