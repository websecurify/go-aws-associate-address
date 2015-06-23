FROM scratch

# ---
# ---
# ---

COPY go-aws-associate-address /

# ---
# ---
# ---

EXPOSE 8080

# ---
# ---
# ---

ENTRYPOINT ["/go-aws-associate-address"]

# ---
