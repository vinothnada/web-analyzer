import { useState } from "react";
import axios from "axios";
import { TextField, Button, Container, Typography, CircularProgress } from "@mui/material";
import { motion } from "framer-motion";

export default function WebAnalyzer() {
  const [url, setUrl] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [results, setResults] = useState(null);

  const isValidUrl = (string) => {
    try {
      const newUrl = new URL(string);
      return newUrl.protocol === "http:" || newUrl.protocol === "https:";
    } catch (_) {
      return false;
    }
  };

  const handleSubmit = async () => {
    if (!isValidUrl(url)) {
      setError("Please enter a valid webpage URL.");
      return;
    }

    setLoading(true);
    setError(null);
    setResults(null);

    try {
      const response = await axios.post("http://localhost:8082/api/analyze", { url });
      setResults(response.data);
    } catch (err) {
      setError(err.response?.data?.message || "Something went wrong");
    }
    setLoading(false);
  };

  return (
    <Container maxWidth="sm" style={{ textAlign: "center", marginTop: "50px" }}>
      <Typography variant="h4" gutterBottom>
        Webpage Analyzer
      </Typography>
      <motion.div initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }}>
        <TextField
          label="Enter URL"
          variant="outlined"
          fullWidth
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          style={{ marginBottom: "20px" }}
          error={!!error}
          helperText={error}
        />
        <Button variant="contained" color="primary" onClick={handleSubmit} disabled={loading}>
          Analyze
        </Button>
      </motion.div>
      {loading && <CircularProgress style={{ marginTop: "20px" }} />}
      {results && (
        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
          <Typography variant="h6">Results:</Typography>
          <Typography>HTML Version: {results.htmlVersion}</Typography>
          <Typography>Title: {results.title}</Typography>
          <Typography>Headings Count: {JSON.stringify(results.headings)}</Typography>
          <Typography>Login Form Present: {results.hasLoginForm ? "Yes" : "No"}</Typography>
          <br/>
          <Typography>Internal Links: {results.internalLinks}</Typography>
          <Typography>External Links: {results.externalLinks}</Typography>
          <Typography>Accesible External Links: {results.accessibleExternalLinks}</Typography>
          <Typography>Broken External Links: {results.brokenExternalLinks}</Typography>
        </motion.div>
      )}
    </Container>
  );
}