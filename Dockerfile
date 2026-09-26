FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
    texlive-latex-recommended texlive-latex-extra \
    texlive-fonts-recommended texlive-fonts-extra \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

ENTRYPOINT ["pdflatex", "-interaction=nonstopmode", "-output-directory=/output"]
