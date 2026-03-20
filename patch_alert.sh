cat << 'DIFF' > patch.diff
<<<<<<< SEARCH
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { CheckCircle2, AlertCircle, AlertTriangle, Search, Filter, MoreHorizontal, Clock, RefreshCw, Activity, Loader2 } from "lucide-react";
=======
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { CheckCircle2, AlertCircle, AlertTriangle, Search, Filter, MoreHorizontal, Clock, RefreshCw, Activity, Loader2, CheckSquare } from "lucide-react";
>>>>>>> REPLACE
<<<<<<< SEARCH
  const [filterSeverity, setFilterSeverity] = useState<string>("all");
  const [filterStatus, setFilterStatus] = useState<string>("all");
  const { toast } = useToast();
=======
  const [filterSeverity, setFilterSeverity] = useState<string>("all");
  const [filterStatus, setFilterStatus] = useState<string>("all");
  const [selectedAlerts, setSelectedAlerts] = useState<string[]>([]);
  const [isBulkUpdating, setIsBulkUpdating] = useState(false);
  const { toast } = useToast();
>>>>>>> REPLACE
<<<<<<< SEARCH
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAlerts();
  }, []);
=======
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAlerts();
  }, []);

  useEffect(() => {
      setSelectedAlerts([]);
  }, [searchQuery, filterSeverity, filterStatus]);
>>>>>>> REPLACE
<<<<<<< SEARCH
        toast({
            title: "Error",
            description: "Failed to update status",
            variant: "destructive",
        });
    }
  };
=======
        toast({
            title: "Error",
            description: "Failed to update status",
            variant: "destructive",
        });
    }
  };

  const handleBulkStatusChange = async (newStatus: AlertStatus) => {
      if (selectedAlerts.length === 0) return;
      setIsBulkUpdating(true);
      try {
          const promises = selectedAlerts.map(id => apiClient.updateAlertStatus(id, newStatus));
          const results = await Promise.all(promises);

          setAlerts(prev => prev.map(a => {
              const updated = results.find(r => r.id === a.id);
              return updated ? updated : a;
          }));

          toast({
              title: "Bulk Update Successful",
              description: `${selectedAlerts.length} alerts marked as ${newStatus}`,
          });
          setSelectedAlerts([]);
      } catch (error) {
          console.error(error);
          toast({
              title: "Error",
              description: "Failed to update some alerts",
              variant: "destructive",
          });
          fetchAlerts(); // Refresh to ensure consistent state
      } finally {
          setIsBulkUpdating(false);
      }
  };

  const toggleSelectAll = () => {
      if (selectedAlerts.length === filteredAlerts.length) {
          setSelectedAlerts([]);
      } else {
          setSelectedAlerts(filteredAlerts.map(a => a.id));
      }
  };

  const toggleSelect = (id: string) => {
      setSelectedAlerts(prev =>
          prev.includes(id) ? prev.filter(aId => aId !== id) : [...prev, id]
      );
  };
>>>>>>> REPLACE
<<<<<<< SEARCH
      <div className="flex flex-col sm:flex-row gap-4 justify-between items-center">
        <div className="relative w-full sm:w-96">
=======
      {selectedAlerts.length > 0 && (
          <div className="bg-muted/50 border rounded-md p-3 flex items-center justify-between mb-4 animate-in slide-in-from-top-2">
              <div className="flex items-center gap-2">
                  <Badge variant="secondary">{selectedAlerts.length} selected</Badge>
              </div>
              <div className="flex items-center gap-2">
                  <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleBulkStatusChange('acknowledged')}
                      disabled={isBulkUpdating}
                  >
                      {isBulkUpdating ? <Loader2 className="h-4 w-4 mr-2 animate-spin" /> : <CheckSquare className="h-4 w-4 mr-2" />}
                      Acknowledge Selected
                  </Button>
                  <Button
                      variant="default"
                      size="sm"
                      onClick={() => handleBulkStatusChange('resolved')}
                      disabled={isBulkUpdating}
                  >
                      {isBulkUpdating ? <Loader2 className="h-4 w-4 mr-2 animate-spin" /> : <CheckCircle2 className="h-4 w-4 mr-2" />}
                      Resolve Selected
                  </Button>
              </div>
          </div>
      )}

      <div className="flex flex-col sm:flex-row gap-4 justify-between items-center">
        <div className="relative w-full sm:w-96">
>>>>>>> REPLACE
<<<<<<< SEARCH
          <TableHeader>
            <TableRow>
              <TableHead className="w-[100px]">Severity</TableHead>
              <TableHead className="w-[100px]">Status</TableHead>
=======
          <TableHeader>
            <TableRow>
              <TableHead className="w-[40px]">
                <Checkbox
                    checked={filteredAlerts.length > 0 && selectedAlerts.length === filteredAlerts.length}
                    onCheckedChange={toggleSelectAll}
                    disabled={filteredAlerts.length === 0}
                    aria-label="Select all"
                />
              </TableHead>
              <TableHead className="w-[100px]">Severity</TableHead>
              <TableHead className="w-[100px]">Status</TableHead>
>>>>>>> REPLACE
<<<<<<< SEARCH
                filteredAlerts.map((alert) => (
                <TableRow key={alert.id} className="group">
                    <TableCell>{getSeverityBadge(alert.severity)}</TableCell>
                    <TableCell>
=======
                filteredAlerts.map((alert) => (
                <TableRow key={alert.id} className="group" data-state={selectedAlerts.includes(alert.id) ? "selected" : undefined}>
                    <TableCell>
                        <Checkbox
                            checked={selectedAlerts.includes(alert.id)}
                            onCheckedChange={() => toggleSelect(alert.id)}
                            aria-label={`Select alert ${alert.id}`}
                        />
                    </TableCell>
                    <TableCell>{getSeverityBadge(alert.severity)}</TableCell>
                    <TableCell>
>>>>>>> REPLACE
DIFF
