import { useState, useEffect, useMemo, useCallback } from 'react';
import { Container, Box, Stepper, Step, StepLabel, Button, Typography, Alert, CircularProgress, Paper } from '@mui/material';
import { loadState, saveState, clearState, type AppState } from './lib/storage';
import { FilePicker } from './components/FilePicker';
import { ParamsForm } from './components/ParamsForm';
import { ProgressBar } from './components/ProgressBar';
import { ScheduleTable } from './components/ScheduleTable';
import { ValidationPanel } from './components/ValidationPanel';
import { DownloadButtons } from './components/DownloadButtons';
import { parseCsv } from './lib/csv';
import type { RunScheduleParams, VersionInfo, ScheduleResponse } from './lib/wasmTypes';

const steps = ['Upload Data', 'Configure Parameters', 'Generate & Review'];

const defaultParams: Partial<RunScheduleParams> = {
  examStartDate: "2025-01-20",
  examEndDate: "2025-01-24",
  slotsPerDay: 2,
  slotDuration: 180,
  tries: 100,
  seed: 0,
  minGap: 60,
  holidays: [],
};

function App() {
  const [appState, setAppState] = useState<AppState>(() => {
    const loaded = loadState();
    if (!loaded.params) loaded.params = defaultParams;
    return loaded;
  });
  const [activeStep, setActiveStep] = useState(0); // Always start at 0 on reload
  const [worker, setWorker] = useState<Worker | null>(null);
  const [workerReady, setWorkerReady] = useState(false);
  const [versionInfo, setVersionInfo] = useState<VersionInfo | null>(null);
  const [generating, setGenerating] = useState(false);
  const [generationResult, setGenerationResult] = useState<ScheduleResponse | null>(appState.lastResult as ScheduleResponse || null);
  const [displayData, setDisplayData] = useState<any[]>([]);

  const scheduleDataPromise = useMemo(() => {
    if (generationResult?.success) {
      return parseCsv(generationResult.scheduleCSV).then(r => r.data);
    }
    return Promise.resolve([]);
  }, [generationResult]);

  useEffect(() => {
    scheduleDataPromise.then(data => setDisplayData(data));
  }, [scheduleDataPromise]);

  useEffect(() => {
    const newWorker = new Worker(new URL('./lib/wasmWorker.ts', import.meta.url), { type: 'module' });
    setWorker(newWorker);

    newWorker.onmessage = (event) => {
      const { type, data } = event.data;
      switch (type) {
        case 'WORKER_READY':
          setWorkerReady(true);
          newWorker.postMessage({ type: 'GET_VERSION' });
          break;
        case 'VERSION_RESULT':
          setVersionInfo(data);
          break;
        case 'RESULT':
        case 'ERROR':
          setGenerationResult(data);
          setGenerating(false);
          updateState({ lastResult: data });
          break;
      }
    };

    return () => {
      newWorker.terminate();
    };
  }, []);

  const updateState = useCallback((newState: Partial<AppState>) => {
    setAppState(prevState => {
      const updatedState = { ...prevState, ...newState };
      saveState(updatedState);
      return updatedState;
    });
  }, []);

  const handleNext = () => {
    const nextStep = activeStep + 1;
    setActiveStep(nextStep);
    updateState({ currentStep: nextStep });
  };

  const handleBack = () => {
    const prevStep = activeStep - 1;
    setActiveStep(prevStep);
    updateState({ currentStep: prevStep });
  };

  const handleReset = () => {
    if (window.confirm("Are you sure you want to clear all data and reset the session?")) {
        clearState();
        setAppState({ params: defaultParams });
        setActiveStep(0);
        setGenerationResult(null);
    }
  };

  const handleGenerate = () => {
    if (!worker || !appState.registrationsFile || !appState.hallsFile || !appState.params) return;

    // Check if column mappings are complete for required fields
    const regMapping = appState.registrationsColumnMapping || {};
    const hallsMapping = appState.hallsColumnMapping || {};

    if (!regMapping.studentIDColumn || !regMapping.courseIDColumn) {
      alert('Please map all required registration columns');
      return;
    }
    if (!hallsMapping.hallIDColumn || !hallsMapping.capacityColumn) {
      alert('Please map all required halls columns');
      return;
    }

    setGenerating(true);
    setGenerationResult(null);

    // Create combined column mapping for the WASM API
    const columnMapping = {
      ...regMapping,
      ...hallsMapping
    };

    // Add columnMapping to params
    const paramsWithMapping = {
      ...appState.params,
      columnMapping
    };

    worker.postMessage({
        type: 'GENERATE_SCHEDULE',
        data: {
            regCSV: appState.registrationsFile.content,
            hallsCSV: appState.hallsFile.content,
            paramsJSON: JSON.stringify(paramsWithMapping),
        }
    });
  };  const isStep1Complete = appState.registrationsFile && appState.hallsFile;
  const isStep2Complete = isStep1Complete && (appState.params as any)?.examStartDate && (appState.params as any)?.examEndDate;

  const getStepContent = (step: number) => {
    switch (step) {
      case 0:
        return (
          <>
            <FilePicker
              title="Registrations CSV"
              file={appState.registrationsFile}
              onFileAccepted={(file) => updateState({ registrationsFile: file })}
              mappingType="registrations"
              columnMapping={appState.registrationsColumnMapping}
              onColumnMappingChange={(mapping) => updateState({ registrationsColumnMapping: mapping })}
            />
            <FilePicker
              title="Halls CSV"
              file={appState.hallsFile}
              onFileAccepted={(file) => updateState({ hallsFile: file })}
              mappingType="halls"
              columnMapping={appState.hallsColumnMapping}
              onColumnMappingChange={(mapping) => updateState({ hallsColumnMapping: mapping })}
            />
          </>
        );
      case 1:
        return (
            <ParamsForm
                params={appState.params || {}}
                setParams={(p: Partial<RunScheduleParams>) => updateState({ params: p })}
            />
        );
      case 2:
        return (
            <>
                <Box sx={{ display: 'flex', justifyContent: 'center', my: 2 }}>
                    <Button
                        variant="contained"
                        color="primary"
                        onClick={handleGenerate}
                        disabled={generating || !workerReady || !isStep2Complete}
                        size="large"
                    >
                        {generating ? <CircularProgress size={24} color="inherit" /> : 'Generate Schedule'}
                    </Button>
                </Box>
                {generating && <ProgressBar message="Generating schedule... this may take a while." />}

                {generationResult && (
                    generationResult.success ? (
                        <>
                            <ValidationPanel report={generationResult.report} />
                            <DownloadButtons
                                scheduleData={displayData}
                                result={generationResult}
                            />
                            <ScheduleTable scheduleData={displayData} />
                        </>
                    ) : (
                        <Alert severity="error" sx={{ mt: 2 }}>
                            <Typography gutterBottom><strong>Scheduling Failed</strong></Typography>
                            {generationResult.error}
                            {generationResult.report && <ValidationPanel report={generationResult.report} />}
                        </Alert>
                    )
                )}
            </>
        );
      default:
        return <Typography>Unknown Step</Typography>;
    }
  };

  return (
    <Container maxWidth="lg">
      <Paper elevation={2} sx={{ my: 4, p: 4 }}>
        <Box sx={{display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start'}}>
            <div>
                <Typography variant="h4" component="h1" gutterBottom>
                Exam Scheduler
                </Typography>
                {versionInfo && <Typography variant="caption" display="block">Powered by {versionInfo.name} v{versionInfo.version} ({versionInfo.go})</Typography>}
                {!workerReady && <Alert severity="info" sx={{mt: 1}}>WASM Worker is loading...</Alert>}
            </div>
            <Button variant="outlined" color="warning" onClick={handleReset}>Reset Session</Button>
        </Box>

        <Stepper activeStep={activeStep} sx={{ my: 4 }}>
          {steps.map((label) => (
            <Step key={label}>
              <StepLabel>{label}</StepLabel>
            </Step>
          ))}
        </Stepper>

        {/* Navigation buttons at the top */}
        <Box sx={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          mb: 3,
          p: 2,
          bgcolor: 'background.default',
          borderRadius: 1,
          border: 1,
          borderColor: 'divider'
        }}>
          <Button
            color="inherit"
            disabled={activeStep === 0}
            onClick={handleBack}
            variant="outlined"
            sx={{ minWidth: 100 }}
          >
            Back
          </Button>

          <Typography variant="body2" color="text.secondary">
            Step {activeStep + 1} of {steps.length}
          </Typography>

          <Button
            onClick={handleNext}
            disabled={activeStep === 1 ? !isStep1Complete : activeStep === 2}
            variant="contained"
            sx={{ minWidth: 100 }}
          >
            Next
          </Button>
        </Box>

        {getStepContent(activeStep)}
      </Paper>
    </Container>
  );
}

export default App;
