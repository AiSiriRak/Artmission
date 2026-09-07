interface DeleteArtworkModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
}

export default function DeleteArtworkModal({ isOpen, onClose, onConfirm }: DeleteArtworkModalProps) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30 backdrop-blur-sm">
      <div className="bg-white rounded-[24px] p-8 max-w-sm w-full mx-4 shadow-xl">
        <h3 className="text-[22px] font-bold text-gray-900 mb-3 text-center leading-tight">
          Are you sure want to<br />delete artwork sample?
        </h3>
        <p className="text-gray-600 text-sm text-center mb-8">
          This artwork will be permanently deleted.<br />This action cannot be undone.
        </p>
        <div className="flex gap-4 justify-center">
          <button 
            className="rounded-[14px] px-6 py-3 border border-gray-300 font-semibold cursor-pointer w-full text-gray-800 hover:bg-gray-50 transition-colors"
            onClick={onClose}
          >
            No, Keep it
          </button>
          <button 
            onClick={onConfirm}
            className="bg-[#FF3333] hover:bg-red-600 text-white rounded-[14px] px-6 py-3 font-semibold cursor-pointer transition-colors w-full"
          >
            Yes, Delete !
          </button>
        </div>
      </div>
    </div>
  );
}