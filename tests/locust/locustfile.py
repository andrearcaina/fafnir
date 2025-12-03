import random
import string

from http_codes import ResponseStatusCodes
from locust import FastHttpUser, between, task
from queries.stock import (
    GET_STOCK_HISTORY,
    GET_STOCK_METADATA,
    GET_STOCK_QUOTE,
    GET_STOCK_QUOTE_BATCH,
    PERIODS,
    SYMBOLS,
)


class FafnirUser(FastHttpUser):
    wait_time = between(1, 3)  # wait between 1 and 3 seconds between tasks

    csrf_token = None

    def on_start(self):
        # perform user registration with random credentials (since each email must be unique)
        email = (
            "email_"
            + "".join(random.choices(string.ascii_lowercase + string.digits, k=8))
            + "@test.com"
        )
        password = "pass_" + "".join(
            random.choices(string.ascii_lowercase + string.digits, k=8)
        )

        registration_data = {
            "email": email,
            "password": password,
            "firstName": "Test",
            "lastName": "User",
        }

        with self.client.post(
            "/auth/register",
            json=registration_data,
            catch_response=True,
        ) as response:
            if response.status_code == ResponseStatusCodes.CREATED.value:
                response.success()
            else:
                response.failure(
                    f"Failed to register user: {response.status_code} - {response.text}"
                )
                return

        # perform user login to obtain auth and CSRF cookies (and CSRF token for future requests)
        login_data = {"email": registration_data["email"], "password": password}
        with self.client.post(
            "/auth/login",
            json=login_data,
            catch_response=True,
        ) as response:
            if response.status_code == ResponseStatusCodes.OK.value:
                for cookie in self.client.cookiejar:
                    if cookie.name == "csrf_token":
                        self.csrf_token = cookie.value

                if not self.csrf_token:
                    response.failure("Auth or CSRF token not found in cookies")
                    return

                response.success()
            else:
                response.failure(
                    f"Failed to login: {response.status_code} - {response.text}"
                )

        return

    def on_stop(self):
        # perform user delete on stop (logout and delete account)
        headers = {
            "X-CSRF-Token": self.csrf_token,
        }

        with self.client.delete(
            "/auth/delete", headers=headers, catch_response=True
        ) as response:
            if response.status_code == ResponseStatusCodes.NO_CONTENT.value:
                response.success()
            else:
                response.failure(
                    f"Failed to logout: {response.status_code} - {response.text}"
                )

        return

    @task(weight=1)
    def get_user_profile(self):
        headers = {
            "X-CSRF-Token": self.csrf_token,
        }

        with self.client.get(
            "/auth/me", headers=headers, catch_response=True
        ) as response:
            if response.status_code == ResponseStatusCodes.OK.value:
                response.success()
            else:
                response.failure(
                    f"Failed to get user profile: {response.status_code} - {response.text}"
                )
                return

    @task(weight=3)
    def get_stock_data(self):
        symbol = random.choice(SYMBOLS)

        with self.client.post(
            "/graphql",
            json={
                "query": GET_STOCK_QUOTE,
                "variables": {"symbol": symbol},
            },
            catch_response=True,
        ) as response:
            if response.status_code == ResponseStatusCodes.OK.value:
                response.success()
            else:
                response.failure(
                    f"Failed to get stock data for {symbol}): {response.status_code} - {response.text}"
                )
                return

    @task(weight=2)
    def get_stock_metadata(self):
        symbol = random.choice(SYMBOLS)

        with self.client.post(
            "/graphql",
            json={"query": GET_STOCK_METADATA, "variables": {"symbol": symbol}},
            catch_response=True,
        ) as response:
            if response.status_code == ResponseStatusCodes.OK.value:
                response.success()
            else:
                response.failure(
                    f"Failed to get stock metadata for {symbol}: {response.status_code} - {response.text}"
                )
                return

    @task(weight=1)
    def get_stock_data_batch(self):
        symbols = random.sample(SYMBOLS, k=3)  # select 3 arbitrary symbols

        with self.client.post(
            "/graphql",
            json={
                "query": GET_STOCK_QUOTE_BATCH,
                "variables": {"symbols": symbols},
            },
            catch_response=True,
        ) as response:
            if response.status_code == ResponseStatusCodes.OK.value:
                response.success()
            else:
                response.failure(
                    f"Failed to get batch stock data for {symbols}): {response.status_code} - {response.text}"
                )
                return

    @task(weight=1)
    def get_stock_history(self):
        symbol = random.choice(SYMBOLS)
        period = random.choice(PERIODS)

        with self.client.post(
            "/graphql",
            json={
                "query": GET_STOCK_HISTORY,
                "variables": {"symbol": symbol, "period": period},
            },
            catch_response=True,
        ) as response:
            if response.status_code == ResponseStatusCodes.OK.value:
                response.success()
            else:
                response.failure(
                    f"Failed to get stock history for {symbol}: {response.status_code} - {response.text}"
                )
                return
