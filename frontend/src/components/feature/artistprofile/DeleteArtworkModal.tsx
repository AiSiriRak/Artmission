import { Button } from "@/components/ui/Button";

interface DeleteArtworkModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
}

export default function DeleteArtworkModal({ 
  isOpen, 
  onClose, 
  onConfirm 
}: DeleteArtworkModalProps) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 backdrop-blur-sm">
      <div className="bg-white rounded-[24px] p-8 max-w-sm w-full mx-4 shadow-xl">
        <h3 className="text-h3 font-bold text-gray-900 mb-3 text-center leading-tight">
          Are you sure want to<br />delete artwork sample?
        </h3>
        <p className="text-gray-600 text-sm text-center mb-8">
          This artwork will be permanently deleted.<br />This action cannot be undone.
        </p>
        
        <div className="flex gap-4 justify-center">
          <Button 
            variant="light" 
            onClick={onClose}
            className="!border"
          >
            No, Keep it
          </Button>
          
          <Button 
            variant="error" 
            onClick={onConfirm}
          >
            Yes, Delete !
          </Button>
        </div>
      </div>
    </div>
  );
}