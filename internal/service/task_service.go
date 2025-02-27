func (s *TaskService) executeTask(task *models.Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, task.Method, task.URL, nil)
	if err != nil {
		return err
	}

	for k, v := range task.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	task.HTTPStatusCode = resp.StatusCode
	task.ResponseHeaders = resp.Header
	task.Length = resp.ContentLength
	task.Status = models.StatusDone

	return nil
}
