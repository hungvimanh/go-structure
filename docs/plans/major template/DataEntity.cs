using CaseExtensions;
using Newtonsoft.Json.Linq;
using SmartFormat;
using System.Diagnostics.CodeAnalysis;
using System.Reflection;

namespace CHECKIN.Common
{
    public class DataEntity : IComparable<DataEntity>
    {
        private static Dictionary<string, JObject> _ErrorResource;
        [JsonIgnore]
        public static Dictionary<string, JObject> ErrorResource
        {
            get
            {
                if (_ErrorResource == null)
                {
                    _ErrorResource = new Dictionary<string, JObject>();
                    List<string> languages = new List<string> { "vi", "en" };
                    foreach (string language in languages)
                    {
                        string folder = Path.Combine("Resources/Errors/" + language);
                        if (Directory.Exists(folder))
                        {
                            List<string> files = Directory.GetFiles(folder, "*.json", SearchOption.AllDirectories).ToList();
                            foreach (string file in files)
                            {
                                string content = File.ReadAllText(file);
                                ErrorResource.Add(language + "." + Path.GetFileNameWithoutExtension(file), JObject.Parse(content));
                            }
                        }
                    }
                }
                return _ErrorResource;
            }
        }

        private string _DataPath;
        private string DataPath
        {
            get
            {
                if (string.IsNullOrEmpty(_DataPath))
                    return GetType().Name;
                else
                    return _DataPath;
            }
            set
            {
                _DataPath = value + "." + GetType().Name;
            }
        }


        private string _BaseLanguage;
        [JsonIgnore]
        public string BaseLanguage
        {
            get
            {
                return _BaseLanguage;
            }
            set
            {
                _BaseLanguage = value;
                List<PropertyInfo> PropertyInfoes = GetType().GetProperties().ToList();
                foreach (PropertyInfo PropertyInfo in PropertyInfoes)
                {
                    if (PropertyInfo.GetGetMethod().IsStatic)
                        continue;
                    if (PropertyInfo.PropertyType.IsGenericType && PropertyInfo.PropertyType.GetGenericTypeDefinition() == typeof(List<>))
                    {
                        IEnumerable<DataEntity> DataEntities = PropertyInfo.GetValue(this) as IEnumerable<DataEntity>;
                        if (DataEntities != null)
                            foreach (DataEntity DataEntity in DataEntities)
                            {
                                DataEntity.DataPath = DataPath;
                                DataEntity.BaseLanguage = _BaseLanguage;
                            }
                    }
                    if (PropertyInfo.PropertyType.IsSubclassOf(typeof(DataEntity)))
                    {
                        DataEntity DataEntity = PropertyInfo.GetValue(this) as DataEntity;
                        if (DataEntity != null)
                        {
                            DataEntity.DataPath = DataPath;
                            DataEntity.BaseLanguage = _BaseLanguage;
                        }
                    }
                }
            }
        }

        private bool _IsValidated = true;
        [JsonIgnore]
        public bool IsValidated
        {
            get
            {
                if (Errors != null && Errors.Count > 0) return _IsValidated = false;
                List<PropertyInfo> PropertiesInfo = GetType().GetProperties().ToList();
                foreach (PropertyInfo PropertyInfo in PropertiesInfo)
                {
                    if (PropertyInfo.PropertyType.IsGenericType && PropertyInfo.PropertyType.GetGenericTypeDefinition() == typeof(List<>))
                    {
                        if (PropertyInfo.GetValue(this) is IEnumerable<DataEntity> DataEntities)
                            foreach (DataEntity DataEntity in DataEntities)
                            {
                                _IsValidated &= DataEntity.IsValidated;
                            }
                    }
                    if (PropertyInfo.PropertyType.IsSubclassOf(typeof(DataEntity)))
                    {
                        if (PropertyInfo.GetValue(this) is DataEntity DataEntity)
                        {
                            _IsValidated &= DataEntity.IsValidated;
                        }
                    }
                }
                return _IsValidated;
            }
        }
        public Dictionary<string, string> Errors { get; set; }
        [JsonIgnore]
        public virtual string Hash { get; set; }
        [JsonIgnore]
        public virtual string Key { get; set; }

        public DataEntity()
        {
        }

        public void AddError(string className, string Key, Enum Value, object parameters = null)
        {
            if (string.IsNullOrEmpty(_BaseLanguage)) _BaseLanguage = "vi";
            if (Errors == null) Errors = new Dictionary<string, string>();

            string file = string.Format("{0}.{1}", _BaseLanguage, className);
            string path = string.Format("{0}.{1}.{2}", DataPath, Key, Value.ToString());
            JToken token = ErrorResource.GetValueOrDefault(file)?.SelectToken(path);
            string content = token == null ? Value.ToString() : token.ToString();
            Key = Key.ToCamelCase();
            if (parameters != null)
                content = Smart.Format(content, parameters);
            if (Errors.ContainsKey(Key))
            {
                if (!Errors[Key].Contains(content))
                    Errors[Key] += content;
            }
            else
                Errors.Add(Key, content);
        }

        public void TrimString()
        {
            List<PropertyInfo> PropertyInfoes = GetType().GetProperties().ToList();
            foreach (PropertyInfo PropertyInfo in PropertyInfoes)
            {
                if (PropertyInfo.PropertyType.IsGenericType && PropertyInfo.PropertyType.GetGenericTypeDefinition() == typeof(List<>))
                {
                    if (PropertyInfo.GetMethod != null)
                    {
                        IEnumerable<DataEntity> DataEntities = PropertyInfo.GetValue(this) as IEnumerable<DataEntity>;
                        if (DataEntities != null)
                            foreach (DataEntity DataEntity in DataEntities)
                            {
                                DataEntity.TrimString();
                            }
                    }
                }
                if (PropertyInfo.PropertyType.IsSubclassOf(typeof(DataEntity)))
                {
                    DataEntity DataEntity = PropertyInfo.GetValue(this) as DataEntity;
                    if (DataEntity != null)
                    {
                        DataEntity.TrimString();
                    }
                }
                if (PropertyInfo.PropertyType.Name == typeof(string).Name)
                {
                    if (PropertyInfo.GetMethod != null && PropertyInfo.SetMethod != null)
                    {
                        string value = PropertyInfo.GetValue(this) as string;
                        value = value?.Trim();
                        PropertyInfo.SetValue(this, value);
                    }
                }
            }
        }

        public virtual int CompareTo([AllowNull] DataEntity other)
        {
            if (other == null)
                return -1;
            if (string.IsNullOrWhiteSpace(Key))
                return -1;
            if (string.IsNullOrWhiteSpace(other.Key))
                return -1;
            return string.Compare(Key, other.Key);
        }
    }
}

