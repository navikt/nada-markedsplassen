import React, { useState, useEffect } from 'react';
import {
    Button,
    TextField,
    Select,
    Alert,
    Table,
    BodyShort,
    Label,
    VStack,
    HStack,
    Checkbox,
    Tooltip
} from '@navikt/ds-react';
import { ExternalLinkIcon, TrashIcon, LinkIcon } from '@navikt/aksel-icons';
import { formatDistanceToNow, addHours, addDays, isAfter } from 'date-fns';
import { nb } from 'date-fns/locale';

export interface TimeRestrictedUrl {
    id: string;
    url: string;
    description: string;
    duration: '12hours' | '1hour';
    createdAt: Date;
    expiresAt: Date;
    isExpired: boolean;
    hasChanges?: boolean;
    editingDescription?: string;
    editingDuration?: '12hours' | '1hour';
}

const TIME_DURATIONS = {
    '1hour': { label: '1 time', value: '1hour', hours: 1 },
    '12hours': { label: '12 timer', value: '12hours', hours: 12 }
} as const;

// Predefined description options in Norwegian
const PREDEFINED_DESCRIPTIONS = [
    'Koderepository',
    'Datakilde',
    'API',
    'Teknisk dokumentasjon',
    'Utviklingsverktøy',
    'Læringsressurser',
    'Feilsporing',
    'Skytjenester',
    'Annet'
];

// Use the same URL validation as the existing URL editor
const isValidUrl = (url: string) => {
    const urlPattern = /^((\*|\*?[a-zA-Z0-9-]+)\.)+[a-zA-Z0-9-]{2,}(\/(\*|[a-zA-Z0-9-._~:/?#[\]@!$&'()*+,;=]*))*$/;
    return !url ? true : urlPattern.test(url);
}

// Mock data for development - using proper URL format without https://
const generateMockData = (): TimeRestrictedUrl[] => [
    {
        id: '1',
        url: 'github.com',
        description: 'Koderepository',
        duration: '1hour',
        createdAt: new Date(Date.now() - 30 * 60 * 1000), // 30 minutes ago
        expiresAt: addHours(new Date(Date.now() - 30 * 60 * 1000), 1),
        isExpired: false
    },
    {
        id: '2',
        url: 'stackoverflow.com',
        description: 'Læringsressurser',
        duration: '12hours',
        createdAt: new Date(Date.now() - 3 * 60 * 60 * 1000), // 3 hours ago
        expiresAt: addHours(new Date(Date.now() - 3 * 60 * 60 * 1000), 12),
        isExpired: false
    },
    {
        id: '3',
        url: 'expired-example.com',
        description: 'Datakilde',
        duration: '1hour',
        createdAt: new Date(Date.now() - 3 * 60 * 60 * 1000), // 3 hours ago
        expiresAt: addHours(new Date(Date.now() - 3 * 60 * 60 * 1000), 1),
        isExpired: true
    },
    {
        id: '4',
        url: 'legacy-system.com',
        description: '', // No description - migrated URL
        duration: '1hour',
        createdAt: new Date(Date.now() - 5 * 60 * 60 * 1000), // 5 hours ago
        expiresAt: addHours(new Date(Date.now() - 5 * 60 * 60 * 1000), 1),
        isExpired: true
    }
];

const TimeRestrictedUrlEditor: React.FC = () => {
    const [timeRestrictedUrls, setTimeRestrictedUrls] = useState<TimeRestrictedUrl[]>([]);
    const [newUrl, setNewUrl] = useState('');
    const [newDescription, setNewDescription] = useState('');
    const [customDescription, setCustomDescription] = useState('');
    const [selectedDuration, setSelectedDuration] = useState<'12hours' | '1hour'>('1hour');
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState<string | null>(null);
    const [selectedUrls, setSelectedUrls] = useState<Set<string>>(new Set());

    // Initialize with mock data
    useEffect(() => {
        setTimeRestrictedUrls(generateMockData());
    }, []);

    // Update expired status periodically
    useEffect(() => {
        const interval = setInterval(() => {
            setTimeRestrictedUrls(prevUrls =>
                prevUrls.map(url => ({
                    ...url,
                    isExpired: isAfter(new Date(), url.expiresAt)
                }))
            );
        }, 60000); // Check every minute

        return () => clearInterval(interval);
    }, []);

    const handleAddUrl = async () => {
        if (!newUrl.trim()) {
            setError('URL kan ikke være tom');
            return;
        }

        if (!isValidUrl(newUrl)) {
            setError('Ugyldig URL format');
            return;
        }

        // Check if this exact URL is already active
        const existingActiveUrl = timeRestrictedUrls.find(url => url.url === newUrl.trim() && !url.isExpired);
        if (existingActiveUrl) {
            setError('Denne URL-en er allerede aktiv. Vent til den utløper eller fjern den først.');
            return;
        }

        setIsLoading(true);
        setError(null);

        try {
            // Simulate API call
            await new Promise(resolve => setTimeout(resolve, 1000));

            const now = new Date();
            const duration = TIME_DURATIONS[selectedDuration];
            const expiresAt = selectedDuration === '1hour'
                ? addHours(now, duration.hours)
                : addHours(now, duration.hours);

            // Use custom description if provided, otherwise use selected predefined description
            const finalDescription = customDescription.trim() || newDescription || 'Ingen beskrivelse';

            const newTimeRestrictedUrl: TimeRestrictedUrl = {
                id: Date.now().toString(),
                url: newUrl.trim(),
                description: finalDescription,
                duration: selectedDuration,
                createdAt: now,
                expiresAt,
                isExpired: false
            };

            setTimeRestrictedUrls(prev => [newTimeRestrictedUrl, ...prev]);
            setNewUrl('');
            setNewDescription('');
            setCustomDescription('');
            setSuccess(`URL lagt til med ${duration.label} tilgang`);

            // Clear success message after 3 seconds
            setTimeout(() => setSuccess(null), 3000);
        } catch (err) {
            setError('Kunne ikke legge til URL');
        } finally {
            setIsLoading(false);
        }
    };

    const handleRemoveUrl = async (id: string) => {
        setIsLoading(true);
        try {
            // Simulate API call
            await new Promise(resolve => setTimeout(resolve, 500));

            setTimeRestrictedUrls(prev => prev.filter(url => url.id !== id));
            setSuccess('URL fjernet');
            setTimeout(() => setSuccess(null), 3000);
        } catch (err) {
            setError('Kunne ikke fjerne URL');
        } finally {
            setIsLoading(false);
        }
    };

    const handleReopenExpiredUrl = async (url: TimeRestrictedUrl) => {
        setIsLoading(true);
        try {
            // Simulate API call
            await new Promise(resolve => setTimeout(resolve, 1000));

            const now = new Date();
            const duration = TIME_DURATIONS[url.duration];
            const expiresAt = url.duration === '12hours'
                ? addHours(now, duration.hours)
                : addDays(now, 1);

            const updatedUrl: TimeRestrictedUrl = {
                ...url,
                createdAt: now,
                expiresAt,
                isExpired: false
            };

            setTimeRestrictedUrls(prev =>
                prev.map(u => u.id === url.id ? updatedUrl : u)
            );

            setSuccess(`URL gjenåpnet med ${duration.label} tilgang`);
            setTimeout(() => setSuccess(null), 3000);
        } catch (err) {
            setError('Kunne ikke gjenåpne URL');
        } finally {
            setIsLoading(false);
        }
    };

    const handleUrlSelection = (urlId: string, checked: boolean) => {
        const newSelectedUrls = new Set(selectedUrls);
        if (checked) {
            newSelectedUrls.add(urlId);
        } else {
            newSelectedUrls.delete(urlId);
        }
        setSelectedUrls(newSelectedUrls);
    };

    const handleOpenSelectedUrls = () => {
        const selectedUrlObjects = timeRestrictedUrls.filter(url =>
            selectedUrls.has(url.id) && url.isExpired
        );

        if (selectedUrlObjects.length === 0) {
            setError('Ingen utløpte URL-er er valgt');
            setTimeout(() => setError(null), 5000);
            return;
        }

        // Open each selected URL in a new tab (no validation needed since checkboxes are disabled for URLs without descriptions)
        selectedUrlObjects.forEach(url => {
            const fullUrl = url.url.startsWith('http') ? url.url : `https://${url.url}`;
            window.open(fullUrl, '_blank', 'noopener,noreferrer');
        });

        setSuccess(`Åpnet ${selectedUrlObjects.length} URL-er i nye faner`);
        setTimeout(() => setSuccess(null), 5000);

        // Clear selection after opening
        setSelectedUrls(new Set());
    };

    const handleOpenSingleUrl = (url: TimeRestrictedUrl) => {
        const fullUrl = url.url.startsWith('http') ? url.url : `https://${url.url}`;
        window.open(fullUrl, '_blank', 'noopener,noreferrer');
    };

    const handleStartEditing = (id: string) => {
        setTimeRestrictedUrls(prev =>
            prev.map(url =>
                url.id === id
                    ? {
                        ...url,
                        hasChanges: true,
                        editingDescription: url.description,
                        editingDuration: url.duration
                    }
                    : url
            )
        );
    };

    const handleCancelEditing = (id: string) => {
        setTimeRestrictedUrls(prev =>
            prev.map(url =>
                url.id === id
                    ? {
                        ...url,
                        hasChanges: false,
                        editingDescription: undefined,
                        editingDuration: undefined
                    }
                    : url
            )
        );
    };

    const handleDescriptionChange = (id: string, value: string) => {
        setTimeRestrictedUrls(prev =>
            prev.map(url => {
                if (url.id === id) {
                    const hasChanges = value !== url.description || (url.editingDuration && url.editingDuration !== url.duration);
                    return {
                        ...url,
                        editingDescription: value,
                        hasChanges: hasChanges
                    };
                }
                return url;
            })
        );
    };

    const handleDurationChange = (id: string, value: '12hours' | '1hour') => {
        setTimeRestrictedUrls(prev =>
            prev.map(url => {
                if (url.id === id) {
                    const hasChanges = value !== url.duration || (url.editingDescription && url.editingDescription !== url.description);
                    return {
                        ...url,
                        editingDuration: value,
                        hasChanges: !!hasChanges
                    };
                }
                return url;
            })
        );
    };

    const handleUpdateUrl = async (id: string) => {
        const url = timeRestrictedUrls.find(u => u.id === id);
        if (!url) return;

        const newDescription = url.editingDescription ?? url.description;
        const newDuration = url.editingDuration ?? url.duration;

        if (!newDescription.trim()) {
            setError('Beskrivelse kan ikke være tom');
            setTimeout(() => setError(null), 3000);
            return;
        }

        setIsLoading(true);
        try {
            // Simulate API call
            await new Promise(resolve => setTimeout(resolve, 1000));

            const now = new Date();
            const durationInfo = TIME_DURATIONS[newDuration];

            // Calculate new expiration time if duration changed
            let newExpiresAt = url.expiresAt;
            if (newDuration !== url.duration && !url.isExpired) {
                newExpiresAt = newDuration === '1hour'
                    ? addHours(now, durationInfo.hours)
                    : addHours(now, durationInfo.hours);
            }

            setTimeRestrictedUrls(prev =>
                prev.map(u =>
                    u.id === id
                        ? {
                            ...u,
                            description: newDescription,
                            duration: newDuration,
                            expiresAt: newExpiresAt,
                            hasChanges: false,
                            editingDescription: undefined,
                            editingDuration: undefined
                        }
                        : u
                )
            );

            setSuccess('URL oppdatert');
            setTimeout(() => setSuccess(null), 3000);
        } catch (err) {
            setError('Kunne ikke oppdatere URL');
            setTimeout(() => setError(null), 3000);
        } finally {
            setIsLoading(false);
        }
    };

    const handleEditingDescriptionChange = (id: string, value: string) => {
        setTimeRestrictedUrls(prev =>
            prev.map(url =>
                url.id === id
                    ? { ...url, editingDescription: value }
                    : url
            )
        );
    };

    const handleEditingDurationChange = (id: string, value: '12hours' | '1hour') => {
        setTimeRestrictedUrls(prev =>
            prev.map(url =>
                url.id === id
                    ? { ...url, editingDuration: value }
                    : url
            )
        );
    };

    const getTimeRemaining = (expiresAt: Date): string => {
        const now = new Date();
        if (isAfter(now, expiresAt)) {
            return 'Utløpt';
        }
        return `Utløper ${formatDistanceToNow(expiresAt, { addSuffix: true, locale: nb })}`;
    };

    const getStatusVariant = (url: TimeRestrictedUrl): 'success' | 'warning' | 'error' => {
        if (url.isExpired) return 'error';

        const timeLeft = url.expiresAt.getTime() - new Date().getTime();
        const oneHour = 60 * 60 * 1000;

        if (timeLeft < oneHour) return 'warning';
        return 'success';
    };

    const expiredUrls = timeRestrictedUrls.filter(url => url.isExpired);

    return (
        <div className="space-y-6">
            <div>
                <Label className="mb-2 block">Legg til tidsbegrenset URL-tilgang</Label>
                <BodyShort className="text-gray-600 mb-4">
                    Legg til URL-er som får midlertidig tilgang. URL-er må være uten https:// prefikset.
                </BodyShort>

                <VStack gap="4" className="max-w-lg">
                    <TextField
                        label="URL"
                        placeholder="example.com"
                        value={newUrl}
                        onChange={(e) => setNewUrl(e.target.value)}
                        error={newUrl && !isValidUrl(newUrl) ? "Ugyldig URL format" : undefined}
                        disabled={isLoading}
                    />

                    <Select
                        label="Beskrivelse"
                        value={newDescription}
                        onChange={(e) => {
                            setNewDescription(e.target.value);
                            // Clear custom description when selecting predefined
                            if (e.target.value && e.target.value !== 'custom') {
                                setCustomDescription('');
                            }
                        }}
                        disabled={isLoading}
                    >
                        <option value="">Velg beskrivelse</option>
                        {PREDEFINED_DESCRIPTIONS.map(desc => (
                            <option key={desc} value={desc}>{desc}</option>
                        ))}
                        <option value="custom">Egendefinert...</option>
                    </Select>

                    {(newDescription === 'custom' || (!PREDEFINED_DESCRIPTIONS.includes(newDescription) && newDescription && newDescription !== 'custom')) && (
                        <TextField
                            label="Egendefinert beskrivelse"
                            placeholder="Skriv din egen beskrivelse"
                            value={customDescription}
                            onChange={(e) => setCustomDescription(e.target.value)}
                            disabled={isLoading}
                        />
                    )}

                    <Select
                        label="Varighet"
                        value={selectedDuration}
                        onChange={(e) => setSelectedDuration(e.target.value as '12hours' | '1hour')}
                        disabled={isLoading}
                    >
                        {Object.values(TIME_DURATIONS).map(duration => (
                            <option key={duration.value} value={duration.value}>
                                {duration.label}
                            </option>
                        ))}
                    </Select>

                    <Button
                        onClick={handleAddUrl}
                        loading={isLoading}
                        disabled={!newUrl.trim() || !isValidUrl(newUrl)}
                        className="self-start"
                    >
                        Legg til URL
                    </Button>
                </VStack>
            </div>

            {error && (
                <Alert variant="error" className="max-w-lg">
                    {error}
                </Alert>
            )}

            {success && (
                <Alert variant="success" className="max-w-lg">
                    {success}
                </Alert>
            )}

            {/* URLs table with checkboxes */}
            {timeRestrictedUrls.length > 0 && (
                <div>
                    <div className="flex items-center justify-between mb-4">
                        <Label>Aktive og tidligere URL-tilganger</Label>
                        {expiredUrls.length > 0 && (
                            <Button
                                variant="secondary"
                                size="small"
                                onClick={handleOpenSelectedUrls}
                                disabled={selectedUrls.size === 0}
                                icon={<LinkIcon aria-hidden />}
                            >
                                Åpne valgte ({selectedUrls.size})
                            </Button>
                        )}
                    </div>

                    <Table className="max-w-6xl">
                        <Table.Header>
                            <Table.Row>
                                <Table.HeaderCell scope="col" className="w-16">Velg</Table.HeaderCell>
                                <Table.HeaderCell scope="col" className="w-64">URL og detaljer</Table.HeaderCell>
                                <Table.HeaderCell scope="col" className="w-40">Status</Table.HeaderCell>
                                <Table.HeaderCell scope="col" className="w-56">Handlinger</Table.HeaderCell>
                            </Table.Row>
                        </Table.Header>
                        <Table.Body>
                            {timeRestrictedUrls
                                .sort((a, b) => {
                                    // Sort by expiry time descending (longest time before expiry first)
                                    return b.expiresAt.getTime() - a.expiresAt.getTime();
                                })
                                .map((url) => (
                                <Table.Row key={url.id}>
                                    <Table.DataCell className="w-16 align-top pt-4">
                                        {(!url.description || url.description.trim() === '') ? (
                                            <Tooltip content="Legg til en beskrivelse for å kunne åpne mot URL-en">
                                                <div>
                                                    <Checkbox
                                                        checked={selectedUrls.has(url.id)}
                                                        onChange={(e) => handleUrlSelection(url.id, e.target.checked)}
                                                        disabled={!url.isExpired || !url.description || url.description.trim() === ''}
                                                        aria-label={`Velg ${url.url}`}
                                                    >
                                                        {/* Empty checkbox content */}
                                                    </Checkbox>
                                                </div>
                                            </Tooltip>
                                        ) : (
                                            <Checkbox
                                                checked={selectedUrls.has(url.id)}
                                                onChange={(e) => handleUrlSelection(url.id, e.target.checked)}
                                                disabled={!url.isExpired || !url.description || url.description.trim() === ''}
                                                aria-label={`Velg ${url.url}`}
                                            >
                                                {/* Empty checkbox content */}
                                            </Checkbox>
                                        )}
                                    </Table.DataCell>
                                    <Table.DataCell className="w-64">
                                        <div className="space-y-2">
                                            <div className="font-mono text-sm font-medium">
                                                {url.url}
                                            </div>
                                            <div className="space-y-1">
                                                <div className="text-xs text-gray-600">Beskrivelse:</div>
                                                <Select
                                                    label=""
                                                    value={url.editingDescription ?? url.description}
                                                    onChange={(e) => handleDescriptionChange(url.id, e.target.value)}
                                                    disabled={isLoading}
                                                    className="w-full"
                                                    size="small"
                                                >
                                                    <option value="">Velg beskrivelse</option>
                                                    {PREDEFINED_DESCRIPTIONS.map(desc => (
                                                        <option key={desc} value={desc}>{desc}</option>
                                                    ))}
                                                    {url.description && !PREDEFINED_DESCRIPTIONS.includes(url.description) && (
                                                        <option value={url.description}>{url.description}</option>
                                                    )}
                                                </Select>
                                            </div>
                                            <div className="space-y-1">
                                                <div className="text-xs text-gray-600">Varighet:</div>
                                                <Select
                                                    label=""
                                                    value={url.editingDuration ?? url.duration}
                                                    onChange={(e) => handleDurationChange(url.id, e.target.value as '12hours' | '1hour')}
                                                    disabled={isLoading}
                                                    className="w-full"
                                                    size="small"
                                                >
                                                    {Object.values(TIME_DURATIONS).map(duration => (
                                                        <option key={duration.value} value={duration.value}>
                                                            {duration.label}
                                                        </option>
                                                    ))}
                                                </Select>
                                            </div>
                                        </div>
                                    </Table.DataCell>
                                    <Table.DataCell className="w-40 align-top pt-4">
                                        {url.isExpired ? (
                                            <span className="text-red-600 text-sm">
                                                Utløpt {formatDistanceToNow(url.expiresAt, { addSuffix: true, locale: nb })}
                                            </span>
                                        ) : (
                                            <span className="text-green-600 text-sm">
                                                Aktiv, utløper {formatDistanceToNow(url.expiresAt, { addSuffix: true, locale: nb })}
                                            </span>
                                        )}
                                    </Table.DataCell>
                                    <Table.DataCell className="w-56 align-top pt-4 h-full min-h-32">
                                        <div className="flex flex-col gap-2 h-full justify-start">
                                            {/* For expired URLs: Åpne, Oppdater, Fjern */}
                                            {url.isExpired ? (
                                                <>
                                                    {/* Åpne button (top for expired URLs) */}
                                                    {(!url.description || url.description.trim() === '') ? (
                                                        <Tooltip content={
                                                            <div className="text-sm font-medium">
                                                                Legg til en beskrivelse for å kunne åpne mot URL-en
                                                            </div>
                                                        }>
                                                            <span className="w-full">
                                                                <Button
                                                                    variant="tertiary"
                                                                    size="small"
                                                                    onClick={() => handleOpenSingleUrl(url)}
                                                                    icon={<ExternalLinkIcon aria-hidden />}
                                                                    disabled={!url.description || url.description.trim() === ''}
                                                                    className="w-full"
                                                                >
                                                                    Åpne
                                                                </Button>
                                                            </span>
                                                        </Tooltip>
                                                    ) : (
                                                        <Button
                                                            variant="tertiary"
                                                            size="small"
                                                            onClick={() => handleOpenSingleUrl(url)}
                                                            icon={<ExternalLinkIcon aria-hidden />}
                                                            disabled={!url.description || url.description.trim() === ''}
                                                            className="w-full"
                                                        >
                                                            Åpne
                                                        </Button>
                                                    )}

                                                    {/* Oppdater button (middle for expired URLs) */}
                                                    <Button
                                                        variant="secondary"
                                                        size="small"
                                                        onClick={() => handleUpdateUrl(url.id)}
                                                        loading={isLoading}
                                                        disabled={!url.hasChanges}
                                                        className="w-full"
                                                    >
                                                        Oppdater
                                                    </Button>

                                                    {/* Fjern button (bottom for expired URLs) */}
                                                    <Button
                                                        variant="danger"
                                                        size="small"
                                                        onClick={() => handleRemoveUrl(url.id)}
                                                        disabled={isLoading}
                                                        icon={<TrashIcon aria-hidden />}
                                                        className="w-full"
                                                    >
                                                        Fjern
                                                    </Button>
                                                </>
                                            ) : (
                                                <>
                                                    {/* For active URLs: Oppdater, Fjern */}

                                                    {/* Oppdater button (top for active URLs) */}
                                                    <Button
                                                        variant="secondary"
                                                        size="small"
                                                        onClick={() => handleUpdateUrl(url.id)}
                                                        loading={isLoading}
                                                        disabled={!url.hasChanges}
                                                        className="w-full"
                                                    >
                                                        Oppdater
                                                    </Button>

                                                    {/* Fjern button (bottom for active URLs) */}
                                                    <Button
                                                        variant="danger"
                                                        size="small"
                                                        onClick={() => handleRemoveUrl(url.id)}
                                                        disabled={isLoading}
                                                        icon={<TrashIcon aria-hidden />}
                                                        className="w-full"
                                                    >
                                                        Fjern
                                                    </Button>
                                                </>
                                            )}
                                        </div>
                                    </Table.DataCell>
                                </Table.Row>
                            ))}
                        </Table.Body>
                    </Table>
                </div>
            )}
        </div>
    );
};

export default TimeRestrictedUrlEditor;
