/**
 * PolicyEditor allows configuring detailed access control policies for a service.
 * It supports defining default actions (Allow/Deny) and a list of specific matching rules
 * based on tool names, arguments, and other criteria.
 */
export function PolicyEditor({ policies = [], onUpdate }: PolicyEditorProps) {
    const { toast } = useToast();
