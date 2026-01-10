
import { GoogleGenAI, Type } from "@google/genai";

// Initialize the Google GenAI client with named parameter and direct process.env.API_KEY access.
const ai = new GoogleGenAI({ apiKey: process.env.API_KEY });

export const parseProblemDescription = async (rawText: string) => {
  // Relying on pre-configured and valid API_KEY as per guidelines.
  const response = await ai.models.generateContent({
    model: 'gemini-3-flash-preview',
    contents: `Parse the following raw practice problem text into structured sections. Respond in JSON.
    Text: ${rawText}`,
    config: {
      responseMimeType: 'application/json',
      responseSchema: {
        type: Type.OBJECT,
        properties: {
          title: { type: Type.STRING },
          description: { type: Type.STRING },
          input: { type: Type.STRING },
          output: { type: Type.STRING },
          constraints: { type: Type.STRING },
          sampleInput: { type: Type.STRING },
          sampleOutput: { type: Type.STRING },
          explanation: { type: Type.STRING }
        },
        required: ['title', 'description']
      }
    }
  });

  try {
    // Correctly using response.text property (not a method).
    return JSON.parse(response.text || '{}');
  } catch (e) {
    console.error("Failed to parse AI response", e);
    return null;
  }
};

export const generateNote = async (problemTitle: string, solutions: any[]) => {
  // Relying on pre-configured and valid API_KEY as per guidelines.
  const response = await ai.models.generateContent({
    model: 'gemini-3-flash-preview',
    contents: `Provide a concise study note or 'Aha!' moment for the following problem and its solution:
    Title: ${problemTitle}
    Solutions: ${JSON.stringify(solutions)}`,
  });

  // response.text is a getter property that returns the string content.
  return response.text;
};