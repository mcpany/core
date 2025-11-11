# Weather.gov API

This directory contains an example of how to use the [weather.gov](https://www.weather.gov/documentation/services-web-api) API.

## Usage

The weather.gov API requires a two-step process to get a weather forecast:

1.  **Get the forecast URL:** Use the `get_grid` call with a latitude and longitude to get a URL for the forecast.
2.  **Get the forecast:** Use the `get_forecast` call with the URL from the previous step to get the weather forecast.

A User-Agent is required to identify your application.
