import { useState } from "react";

export const usePhoneInput = () => {
  const [phone, setPhone] = useState("");

  const formatPhone = (digits) => {
    let formatted = "+7";
    if (digits.length > 1) {
      const phoneDigits = digits.slice(1);
      formatted += ` (${phoneDigits.slice(0, 3)}`;
      if (phoneDigits.length > 3) {
        formatted += `) ${phoneDigits.slice(3, 6)}`;
        if (phoneDigits.length > 6) {
          formatted += `-${phoneDigits.slice(6, 8)}`;
          if (phoneDigits.length > 8) {
            formatted += `-${phoneDigits.slice(8, 10)}`;
          }
        }
      }
    }
    return formatted;
  };

  const handlePhoneChange = (e) => {
    const value = e.target.value;
    let digits = value.replace(/\D/g, "");

    if (!digits.startsWith("7") && !digits.startsWith("8")) {
      digits = "7" + digits;
    } else {
      digits = digits.replace(/^[78]/, "7");
    }

    digits = digits.slice(0, 11);
    setPhone(digits);
  };

  const displayValue = phone ? formatPhone(phone) : "";

  return {
    phone,
    displayValue,
    handlePhoneChange,
    isPhoneValid: phone.length === 11,
    reset: () => setPhone(""),
  };
};
