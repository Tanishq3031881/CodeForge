FROM node:22-alpine
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm install --omit=dev
COPY server.js persistence.js* ./
ENV HOST=0.0.0.0
EXPOSE 1234
USER node
CMD ["node", "server.js"]
