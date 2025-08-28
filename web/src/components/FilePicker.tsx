import React, { useCallback, useState, useEffect } from 'react';
import { useDropzone } from 'react-dropzone';
import { Box, Typography, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Alert, FormControl, InputLabel, Select, MenuItem, Card, CardContent, Chip } from '@mui/material';
import type { SelectChangeEvent } from '@mui/material';
import { parseCsv } from '../lib/csv';
import type { ColumnMapping } from '../lib/wasmTypes';

interface FilePickerProps {
  title: string;
  file: { name: string; content: string } | undefined;
  onFileAccepted: (file: { name: string; content: string }) => void;
  mappingType: 'registrations' | 'halls'; // Determines which fields to show
  columnMapping?: ColumnMapping;
  onColumnMappingChange?: (mapping: ColumnMapping) => void;
}

export const FilePicker: React.FC<FilePickerProps> = ({
  title,
  file,
  onFileAccepted,
  mappingType,
  columnMapping = {},
  onColumnMappingChange
}) => {
  const [preview, setPreview] = useState<any[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [availableColumns, setAvailableColumns] = useState<string[]>([]);
  const [currentMapping, setCurrentMapping] = useState<ColumnMapping>(columnMapping || {});

  // Sync currentMapping with columnMapping prop when it changes
  useEffect(() => {
    setCurrentMapping(columnMapping || {});
  }, [columnMapping]);

  // Define fields based on mapping type
  const getRequiredFields = () => {
    if (mappingType === 'registrations') {
      return [
        { key: 'studentIDColumn', label: 'Student ID Column', defaultName: 'student_id' },
        { key: 'courseIDColumn', label: 'Course ID Column', defaultName: 'course_id' }
      ];
    } else {
      return [
        { key: 'hallIDColumn', label: 'Hall ID Column', defaultName: 'hall' },
        { key: 'capacityColumn', label: 'Capacity Column', defaultName: 'capacity' },
        { key: 'groupColumn', label: 'Group Column (Optional)', defaultName: 'group' }
      ];
    }
  };

  const requiredFields = getRequiredFields();

  const processFile = useCallback(async (fileContent: string) => {
    try {
      const parseResult = await parseCsv(fileContent);
      if (parseResult.errors.length > 0) {
        setError(`CSV parsing error: ${parseResult.errors[0].message}`);
        setPreview([]);
        setAvailableColumns([]);
      } else {
        const columns = parseResult.meta.fields || [];
        setAvailableColumns(columns);

        // Auto-map columns if they match exactly
        const autoMapping: ColumnMapping = {};
        requiredFields.forEach(field => {
          const exactMatch = columns.find(col => col.toLowerCase() === field.defaultName.toLowerCase());
          if (exactMatch) {
            (autoMapping as any)[field.key] = exactMatch;
          }
        });

        // Only update mapping if we found auto-mappings and current mapping is empty
        setCurrentMapping(prev => {
          const hasExistingMappings = Object.values(prev).some(val => val);
          if (hasExistingMappings) {
            return prev; // Don't override existing mappings
          }
          return { ...prev, ...autoMapping };
        });

        // Check for missing required mappings (excluding optional fields)
        const requiredMappings = requiredFields.filter(f => !f.label.includes('Optional'));
        const mappingToCheck = Object.values(columnMapping).some(val => val) ? columnMapping : autoMapping;
        const missingMappings = requiredMappings.filter(field => !(mappingToCheck as any)[field.key]);

        if (missingMappings.length > 0) {
          setError(`Please map the following required fields: ${missingMappings.map(f => f.label).join(', ')}`);
        } else {
          setError(null);
        }

        setPreview(parseResult.data.slice(0, 5));
      }
    } catch (e: any) {
      setError(`Failed to parse CSV: ${e.message}`);
      setPreview([]);
      setAvailableColumns([]);
    }
  }, [requiredFields, columnMapping]);

  useEffect(() => {
    if (file?.content) {
      processFile(file.content);
    }
  }, [file?.content, file?.name]); // Only depend on content and name, not the processFile function

  useEffect(() => {
    if (onColumnMappingChange && JSON.stringify(currentMapping) !== JSON.stringify(columnMapping)) {
      onColumnMappingChange(currentMapping);
    }
  }, [currentMapping]); // Remove onColumnMappingChange from dependencies to prevent infinite loop

  const handleMappingChange = (fieldKey: keyof ColumnMapping) => (event: SelectChangeEvent) => {
    const newMapping: ColumnMapping = {
      ...currentMapping,
      [fieldKey]: event.target.value
    };
    setCurrentMapping(newMapping);

    // Check if all required fields are now mapped
    const requiredMappings = requiredFields.filter(f => !f.label.includes('Optional'));
    const missingMappings = requiredMappings.filter(field => !(newMapping as any)[field.key]);
    if (missingMappings.length === 0) {
      setError(null);
    } else {
      setError(`Please map the following required fields: ${missingMappings.map(f => f.label).join(', ')}`);
    }
  };  const onDrop = useCallback(
    (acceptedFiles: File[]) => {
      if (acceptedFiles.length > 0) {
        const droppedFile = acceptedFiles[0];
        const reader = new FileReader();
        reader.onload = async () => {
          const content = reader.result as string;
          onFileAccepted({ name: droppedFile.name, content });
        };
        reader.readAsText(droppedFile);
      }
    },
    [onFileAccepted]
  );

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: { 'text/csv': ['.csv'] },
    maxFiles: 1,
  });

  return (
    <Card elevation={1} sx={{ mt: 2, bgcolor: 'background.paper' }}>
      <CardContent>
        <Typography variant="h6" sx={{ mb: 2, color: 'primary.main', fontWeight: 600 }}>
          📄 {title}
        </Typography>
        <Box
          {...getRootProps()}
          sx={{
            border: '2px dashed',
            borderColor: isDragActive ? 'primary.main' : 'divider',
            borderRadius: 2,
            p: 4,
            textAlign: 'center',
            cursor: 'pointer',
            backgroundColor: isDragActive ? 'action.hover' : 'background.default',
            transition: 'all 0.2s ease-in-out',
            '&:hover': {
              borderColor: 'primary.main',
              backgroundColor: 'action.hover',
              boxShadow: 1,
            },
          }}
        >
          <input {...getInputProps()} />
          {file ? (
            <Box>
              <Typography variant="h6" color="success.main" sx={{ mb: 1 }}>
                ✅ {file.name}
              </Typography>
              <Chip label="File loaded" color="success" size="small" />
            </Box>
          ) : (
            <Box>
              <Typography variant="h6" color="text.secondary" sx={{ mb: 1 }}>
                📁 Drop your CSV file here
              </Typography>
              <Typography variant="body2" color="text.secondary">
                or click to browse files
              </Typography>
            </Box>
          )}
        </Box>

        {availableColumns.length > 0 && (
          <Box sx={{ mt: 3 }}>
            <Typography variant="h6" gutterBottom sx={{ color: 'primary.main', fontWeight: 600 }}>
              🔗 Column Mapping
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
              Map your CSV columns to the required fields
            </Typography>
            <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: 2 }}>
              {requiredFields.map(field => (
                <FormControl key={field.key} fullWidth>
                  <InputLabel>{field.label}</InputLabel>
                  <Select
                    value={(currentMapping as any)[field.key] || ''}
                    label={field.label}
                    onChange={handleMappingChange(field.key as keyof ColumnMapping)}
                    sx={{
                      borderRadius: 1,
                    }}
                  >
                    <MenuItem value="">
                      <em>Select column...</em>
                    </MenuItem>
                    {availableColumns.map(col => (
                      <MenuItem key={col} value={col}>{col}</MenuItem>
                    ))}
                  </Select>
                </FormControl>
              ))}
            </Box>
          </Box>
        )}

        {error && (
          <Alert severity="error" sx={{ mt: 2 }}>
            {error}
          </Alert>
        )}

        {preview.length > 0 && (
          <Box sx={{ mt: 3 }}>
            <Typography variant="h6" gutterBottom sx={{ color: 'primary.main', fontWeight: 600 }}>
              👁️ Data Preview
            </Typography>
            <TableContainer component={Paper}>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    {Object.keys(preview[0] || {}).map(key => (
                      <TableCell key={key} sx={{ fontWeight: 600 }}>
                        {key}
                        {/* Show which required field this column is mapped to */}
                        {Object.entries(currentMapping).find(([_, col]) => col === key) && (
                          <Typography variant="caption" color="primary" display="block">
                            → {Object.entries(currentMapping).find(([_, col]) => col === key)?.[0]}
                          </Typography>
                        )}
                      </TableCell>
                    ))}
                  </TableRow>
                </TableHead>
                <TableBody>
                  {preview.map((row, index) => (
                    <TableRow key={index}>
                      {Object.values(row).map((value: any, i) => (
                        <TableCell key={i}>{value}</TableCell>
                      ))}
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          </Box>
        )}
      </CardContent>
    </Card>
  );
};
