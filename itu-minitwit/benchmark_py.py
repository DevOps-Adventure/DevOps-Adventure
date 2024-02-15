import time

def is_prime(n):
    if n <= 1:
        return False
    if n <= 3:
        return True
    if n % 2 == 0 or n % 3 == 0:
        return False
    i = 5
    while i * i <= n:
        if n % i == 0 or n % (i + 2) == 0:
            return False
        i += 6
    return True

def calculate_primes(max_number):
    primes = []
    for number in range(2, max_number):
        if is_prime(number):
            primes.append(number)
    return primes

# Measure execution time
start_time = time.time()
primes = calculate_primes(50000)  # Find primes up to 50,000
end_time = time.time()

print("Number of primes found:", len(primes))
print("Time taken in Python:", end_time - start_time, "seconds")
