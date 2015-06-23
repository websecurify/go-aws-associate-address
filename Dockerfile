FROM scratch

# ---
# ---
# ---

COPY go-aws-associate-address /

# ---
# ---
# ---

ENTRYPOINT ["/go-aws-associate-address"]

# ---
