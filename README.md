# web-analyzer

Please find the steps to run the WebAnalyzer project

Run Go server
--------------
1. Clone the project
2. Navigate into web-analyzer/src/
3. Run "go mod tidy"
4. Run "go run main.go"


Run UI
------------
1. Navigate into web-analyzer/web-analyzer-ui
2. npm install
3. npm run dev

Potential feature improvements (can be added in future)
-----------------------

1. Caching : Consider caching the results of analysis for pages that are frequently checked. This can reduce server load and provide faster results for subsequent users.
2. Rate Limiting: Implement rate limiting to prevent abuse, especially when handling external links.
3. Asynchronous Processing: For very large pages, consider using background processing or queueing, so users can check back later instead of waiting for a long time.
4. Network Issues: Handle scenarios where a network issue occurs (like timeouts) and provide the user with relevant messages.
5. Custom Error Pages: Design custom error pages to explain what went wrong, e.g., "The server could not process your request due to a timeout" with a suggestion to try again later.
6. History of Analysis: Provide users with an option to save previous analysis results, perhaps linking it to their account. Users could review their history or download reports.
7. Email Reports: Allow users to send themselves a report of the analysis in PDF or CSV format for future reference.
8. Integration with Google Analytics or Search Console: If possible, allow users to fetch data from their Google Analytics.
9. Crawler Integration: Integrate with existing web crawlers to get deeper analysis of complex pages (e.g., dynamic content, JavaScript-rendered pages).
10. Export Data: Let users export the results into CSV, PDF, or JSON formats for further analysis.