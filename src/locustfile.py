"""
Locust load test for the Product API.

Run with:
  locust -f locustfile.py --host=http://<YOUR_ALB_DNS_OR_LOCALHOST:8080>

Then open http://localhost:8089 in your browser to start the test.
"""

import random
import string
from locust import HttpUser, FastHttpUser, task, between, events


def random_sku(length=10):
    """Generate a random SKU string."""
    return "".join(random.choices(string.ascii_uppercase + string.digits, k=length))


def random_product(pid):
    """Generate a valid product payload."""
    return {
        "product_id": pid,
        "sku": random_sku(),
        "manufacturer": f"Manufacturer-{random.randint(1, 100)}",
        "category_id": random.randint(1, 50),
        "weight": random.randint(0, 10000),
        "some_other_id": random.randint(1, 1000),
    }


# ──────────────────────────────────────────────
# Test with standard HttpUser
# ──────────────────────────────────────────────
class ProductHttpUser(HttpUser):
    """Standard HttpUser - uses Python requests library under the hood."""

    wait_time = between(1, 3)

    # Track the highest product ID we've created so GETs have valid targets
    max_product_id = 0

    @task(1)  # weight=1 → less frequent
    def create_product(self):
        """POST a new product (write operation)."""
        ProductHttpUser.max_product_id += 1
        pid = ProductHttpUser.max_product_id
        payload = random_product(pid)
        self.client.post(
            f"/products/{pid}/details",
            json=payload,
            name="/products/[id]/details",
        )

    @task(5)  # weight=5 → more frequent (simulates read-heavy real world)
    def get_product(self):
        """GET an existing product (read operation)."""
        if ProductHttpUser.max_product_id < 1:
            # No products yet, create one first
            self.create_product()
            return
        pid = random.randint(1, ProductHttpUser.max_product_id)
        self.client.get(
            f"/products/{pid}",
            name="/products/[id]",
        )

    @task(1)
    def get_nonexistent_product(self):
        """GET a product that doesn't exist (test 404 handling)."""
        pid = random.randint(900000, 999999)
        with self.client.get(
            f"/products/{pid}",
            name="/products/[id] (404)",
            catch_response=True,
        ) as resp:
            if resp.status_code == 404:
                resp.success()  # 404 is expected here, don't count as failure

    @task(1)
    def post_invalid_product(self):
        """POST invalid data (test 400 handling)."""
        pid = random.randint(1, 100)
        with self.client.post(
            f"/products/{pid}/details",
            json={"product_id": pid},  # missing required fields
            name="/products/[id]/details (400)",
            catch_response=True,
        ) as resp:
            if resp.status_code == 400:
                resp.success()


# ──────────────────────────────────────────────
# Test with FastHttpUser for comparison
# ──────────────────────────────────────────────
class ProductFastHttpUser(FastHttpUser):
    """FastHttpUser - uses geventhttpclient, lower overhead per request."""

    wait_time = between(1, 3)
    max_product_id = 0

    @task(1)
    def create_product(self):
        ProductFastHttpUser.max_product_id += 1
        pid = ProductFastHttpUser.max_product_id
        payload = random_product(pid)
        self.client.post(
            f"/products/{pid}/details",
            json=payload,
            name="/products/[id]/details",
        )

    @task(5)
    def get_product(self):
        if ProductFastHttpUser.max_product_id < 1:
            self.create_product()
            return
        pid = random.randint(1, ProductFastHttpUser.max_product_id)
        self.client.get(
            f"/products/{pid}",
            name="/products/[id]",
        )

    @task(1)
    def get_nonexistent_product(self):
        pid = random.randint(900000, 999999)
        with self.client.get(
            f"/products/{pid}",
            name="/products/[id] (404)",
            catch_response=True,
        ) as resp:
            if resp.status_code == 404:
                resp.success()