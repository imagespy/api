Feature: API Endpoint /v2/images/{reference}/children
  V2 Image Children API Endpoint
  Scenario: Children of image that has children
    Given the Docker image "127.0.0.1:52854/golang:latest" pushed
    And the Docker image "127.0.0.1:52854/debian:stretch-20190326" pushed
    And imagespy API started
    When sending a "POST" request to "/v2/images/127.0.0.1:52854/golang:latest"
    And sending a "POST" request to "/v2/images/127.0.0.1:52854/debian:stretch-20190326"
    And sending a "GET" request to "/v2/images/127.0.0.1:52854/debian:stretch-20190326/children"
    Then the API responds with status code "200"
    And the API responds with a body of "fixtures/results/g.json"

  Scenario: Children of image that has no children
    Given the Docker image "127.0.0.1:52854/debian:stretch-20190326" pushed
    And imagespy API started
    When sending a "POST" request to "/v2/images/127.0.0.1:52854/debian:stretch-20190326"
    And sending a "GET" request to "/v2/images/127.0.0.1:52854/debian:stretch-20190326/children"
    Then the API responds with status code "200"
    And the API responds with a body of "fixtures/results/h.json"
